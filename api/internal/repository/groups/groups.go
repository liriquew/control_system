package groups_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/liriquew/control_system/internal/entities"
	"github.com/liriquew/control_system/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type GroupsRepository struct {
	db *sqlx.DB
}

func NewGroupsRepository(db *sqlx.DB) (*GroupsRepository, error) {
	return &GroupsRepository{
		db: db,
	}, nil
}

var (
	ErrDenied = errors.New("access denied")

	ErrNotFound        = errors.New("not found")
	ErrNotExists       = errors.New("not exists")
	ErrNothingToUpdate = errors.New("nothing to update (empty updatable fields)")
	ErrInvalideRole    = errors.New("invalide role param")
	ErrWrongOwnerID    = errors.New("user with owner id not found")
)

func (gr *GroupsRepository) CheckAccess(ctx context.Context, userID, groupID int64) error {
	query := `
		SELECT 1 FROM group_members gm
		JOIN groups g ON g.id = gm.group_id
		JOIN users u ON u.id = gm.user_id
		WHERE u.id = $1 AND g.id = $2 
	`

	var val int
	if err := gr.db.QueryRowContext(ctx, query, userID, groupID).Scan(&val); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrDenied
		}
		return err
	}

	return nil
}

func (gr *GroupsRepository) checkPermission(ctx context.Context, userID, groupID int64, roles []string) error {
	query := `
		SELECT 1 FROM group_members gm 
		JOIN groups g ON g.id = gm.group_id 
		JOIN users u ON u.id = gm.user_id
		WHERE u.id = $1 AND g.id = $2 AND gm.role = ANY($3)
	`

	var val int64
	err := gr.db.QueryRowContext(ctx, query, userID, groupID, pq.Array(roles)).Scan(&val)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrDenied
		}

		return fmt.Errorf("checkPermission error: %w", err)
	}

	return nil
}

func (gr *GroupsRepository) CheckEditorPermission(ctx context.Context, userID, groupID int64) error {
	return gr.checkPermission(ctx, userID, groupID, []string{"admin", "editor"})
}

func (gr *GroupsRepository) CheckAdminPermission(ctx context.Context, userID, groupID int64) error {
	return gr.checkPermission(ctx, userID, groupID, []string{"admin"})
}

func (gr *GroupsRepository) Begin(ctx context.Context) (*sql.Tx, error) {
	return gr.db.BeginTx(ctx, nil)
}

func (gr *GroupsRepository) CreateGroup(ctx context.Context, group *models.Group) (*models.Group, error) {
	query := "INSERT INTO groups (owner_id, name, description) VALUES ($1, $2, $3) RETURNING id"

	txn, err := gr.db.Begin()
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	err = txn.QueryRowContext(ctx, query, group.OwnerID, group.Name, group.Description).Scan(&group.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // Код ошибки для FOREIGN KEY violation
				switch pqErr.Constraint {
				case "fk_group_owner":
					return nil, fmt.Errorf("user with id %d does not exists%w", group.OwnerID, ErrWrongOwnerID)
				}
			default:
				return nil, fmt.Errorf("database error: %s", pqErr.Message)
			}
		}
		return nil, err
	}

	query = "INSERT INTO group_members (user_id, group_id, role) VALUES ($1, $2, $3)"
	if _, err := txn.ExecContext(ctx, query, group.OwnerID, group.ID, "admin"); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // Код ошибки для FOREIGN KEY violation
				switch pqErr.Constraint {
				case "fk_member_group":
					return nil, fmt.Errorf("group with id %d does not exist%w", group.ID, ErrNotExists)
				case "fk_member_user":
					return nil, fmt.Errorf("user with id %d does not exist%w", group.OwnerID, ErrNotExists)
				}
			case "23514": // Код ошибки для CHECK violation
				if pqErr.Constraint == "no_group_member_role" {
					return nil, ErrInvalideRole
				}
			default:
				return nil, fmt.Errorf("database error: %s", pqErr.Message)
			}
		}
	}

	if err := txn.Commit(); err != nil {
		return nil, err
	}

	return group, nil
}

func (gr *GroupsRepository) GetGroup(ctx context.Context, userID, groupID int64) (*models.Group, error) {
	query := `
		SELECT g.id, g.owner_id, g.name, g.description, g.created_at FROM groups g
		JOIN group_members m ON g.id = m.group_id
		WHERE g.id = $1 AND m.user_id = $2
	`

	var group models.Group
	err := gr.db.GetContext(ctx, &group, query, groupID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &group, err
}

func (gr *GroupsRepository) ListUserGroups(ctx context.Context, userID int64) ([]*models.Group, error) {
	query := `
		SELECT * FROM groups
		WHERE id IN (SELECT group_id FROM group_members WHERE user_id=$1)
	`

	var groups []*models.Group
	err := gr.db.SelectContext(ctx, &groups, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("groups not found%w", ErrNotFound)
		}
		return nil, err
	}

	return groups, err
}

func (gr *GroupsRepository) DeleteGroup(ctx context.Context, ownerID, groupID int64) (bool, error) {
	query := `DELETE FROM groups WHERE owner_id = $1 AND id = $2 `

	result, err := gr.db.ExecContext(ctx, query, ownerID, groupID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, ErrNotFound
	}

	return rowsAffected == 0, nil
}

func (gr *GroupsRepository) UpdateGroup(ctx context.Context, group *models.Group) error {
	var fields []string
	args := []any{
		0: group.ID,
		1: group.OwnerID,
	}
	argPos := 3

	if group.Name != "" {
		fields = append(fields, fmt.Sprintf("name=$%d", argPos))
		args = append(args, group.Name)
		argPos++
	}

	if group.Description != "" {
		fields = append(fields, fmt.Sprintf("description=$%d", argPos))
		args = append(args, group.Description)
		argPos++
	}

	if len(fields) == 0 {
		return ErrNothingToUpdate
	}

	query := fmt.Sprintf("UPDATE groups SET %s WHERE id=$1 AND owner_id=$2", strings.Join(fields, ", "))

	result, err := gr.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (gr *GroupsRepository) ListGroupMembers(ctx context.Context, userID, groupID int64) ([]*models.User, error) {
	query := `
        SELECT u.id, u.username FROM group_members gm
        JOIN users u ON gm.user_id = u.id
        WHERE gm.group_id = $2
        AND EXISTS (
            SELECT 1 FROM group_members 
            WHERE user_id = $1 AND group_id = $2
		)
	`

	var users []*models.User
	err := gr.db.SelectContext(ctx, &users, query, userID, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return users, nil
}

func (gr *GroupsRepository) AddGroupMember(ctx context.Context, ownerID int64, member *models.GroupMember) (*models.GroupMember, error) {
	if member.Role == "" {
		member.Role = "member"
	}

	validRoles := map[string]struct{}{
		"admin":  {},
		"editor": {},
		"member": {},
		"viewer": {},
	}
	if _, ok := validRoles[member.Role]; !ok {
		return nil, ErrInvalideRole
	}

	query := `
        INSERT INTO group_members (group_id, user_id, role)
        VALUES ($1, $2, $3)
    `

	_, err := gr.db.ExecContext(ctx, query, member.GroupID, member.UserID, member.Role)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // Код ошибки для FOREIGN KEY violation
				switch pqErr.Constraint {
				case "fk_member_group":
					return nil, fmt.Errorf("group with id %d does not exist%w", member.GroupID, ErrNotExists)
				case "fk_member_user":
					return nil, fmt.Errorf("user with id %d does not exist%w", member.UserID, ErrNotExists)
				}
			case "23514": // Код ошибки для CHECK violation
				if pqErr.Constraint == "no_group_member_role" {
					return nil, ErrInvalideRole
				}
			default:
				return nil, fmt.Errorf("database error: %s", pqErr.Message)
			}
		}
		return nil, err
	}

	return member, nil
}

func (gr *GroupsRepository) RemoveGroupMember(ctx context.Context, ownerID int64, member *models.GroupMember) (bool, error) {
	query := `
		DELETE FROM group_members 
		WHERE 
			group_id = $1 AND 
			user_id = $2 AND 
			(
			EXISTS (
				SELECT 1 FROM groups 
				WHERE id = $1 AND owner_id = $3
			)
		)
	`

	result, err := gr.db.ExecContext(ctx, query, member.GroupID, member.UserID, ownerID)
	if err != nil {
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowsAffected == 0 {
		return true, ErrNotFound
	}

	return true, nil
}

func (gr *GroupsRepository) ChangeMemberRole(ctx context.Context, ownerID int64, member *models.GroupMember) error {
	validRoles := map[string]struct{}{
		"admin":  {},
		"editor": {},
		"member": {},
		"viewer": {},
	}
	if _, ok := validRoles[member.Role]; !ok {
		return ErrInvalideRole
	}

	query := `
        UPDATE group_members SET role = $1 WHERE group_id = $2 AND user_id = $3 AND (
			EXISTS (
				SELECT 1 FROM groups 
				WHERE id = $2 AND owner_id = $4
			)
		)
    `

	result, err := gr.db.ExecContext(ctx, query, member.Role, member.GroupID, member.UserID, ownerID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (gr *GroupsRepository) CreateGroupGraph(ctx context.Context, graph *models.Graph, nodes []*models.Node) (*models.Graph, error) {
	txn, err := gr.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	graphQuery := `INSERT INTO graphs (group_id, created_by, name) VALUES ($1, $2, $3) RETURNING id`
	err = txn.QueryRowContext(ctx, graphQuery, graph.GroupID, graph.CreatedBy, graph.Name).Scan(&graph.ID)
	if err != nil {
		return nil, fmt.Errorf("error while creating graph: %w", err)
	}

	// Узлы
	// TODO: use prepare statement
	nodeQuery := `INSERT INTO nodes (graph_id, task_id%s) VALUES ($1, $2%s) RETURNING id`
	nodeIDMap := make(map[int64]int64, len(nodes))
	for _, node := range nodes {
		var currentQuery string
		args := []any{
			0: graph.ID,
			1: node.TaskID,
		}
		if node.AssignedTo != nil {
			currentQuery = fmt.Sprintf(nodeQuery, ", assigned_to", ", $3")
			args = append(args, *node.AssignedTo)
		} else {
			currentQuery = fmt.Sprintf(nodeQuery, "", "")
		}

		var nodeID int64
		err = txn.QueryRowContext(ctx, currentQuery, args...).Scan(&nodeID)
		if err != nil {
			return nil, fmt.Errorf("error while creating node: %w", err)
		}
		nodeIDMap[node.ID] = nodeID
		fmt.Println(node, node.DependencyNodeIDs)
		node.ID = nodeID
	}

	// Зависимости
	dependencyQuery := `INSERT INTO dependencies (from_node_id, to_node_id, graph_id) VALUES ($1, $2, $3)`
	stmt, err := txn.PrepareContext(ctx, dependencyQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	for _, node := range nodes {
		fromNodeID := node.ID
		for _, dependencyNodeDummyID := range node.DependencyNodeIDs {
			toNodeID := nodeIDMap[dependencyNodeDummyID]
			fmt.Println(fromNodeID, toNodeID, graph.ID, dependencyNodeDummyID, nodeIDMap)
			_, err = stmt.ExecContext(ctx, fromNodeID, toNodeID, graph.ID)
			if err != nil {
				return nil, fmt.Errorf("error while creating dependency: %w", err)
			}
			node.DependencyNodeIDs = append(node.DependencyNodeIDs, toNodeID)
		}
	}

	if err = txn.Commit(); err != nil {
		return nil, err
	}

	return graph, nil
}

func (gr *GroupsRepository) ListGroupGraphs(ctx context.Context, userID int64, groupID int64) ([]*entities.GraphWithNodes, error) {
	query := `
		SELECT * FROM graphs WHERE group_id=$1 AND EXISTS (
            SELECT 1 FROM group_members 
            WHERE group_id=$1 AND user_id=$2
        )
	`

	var graphs []models.Graph
	err := gr.db.SelectContext(ctx, &graphs, query, groupID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error while getting graphs: %w", err)
	}

	graphsWithNodes := make([]*entities.GraphWithNodes, len(graphs))
	for i := range graphs {
		graphsWithNodes[i] = &entities.GraphWithNodes{
			GraphInfo: graphs[i],
		}
	}

	queryNodes := "SELECT * FROM nodes WHERE graph_id=$1"
	queryDependencies := "SELECT to_node_id FROM dependencies WHERE from_node_id=$1"
	for i, graph := range graphs {
		if err := gr.db.SelectContext(ctx, &graphsWithNodes[i].Nodes, queryNodes, graph.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return nil, fmt.Errorf("error while getting nodes: %w", err)
		}

		for _, node := range graphsWithNodes[i].Nodes {
			if err := gr.db.SelectContext(ctx, &node.DependencyNodeIDs, queryDependencies, node.ID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}

				return nil, fmt.Errorf("error while getting dependencies: %w", err)
			}
		}
	}

	return graphsWithNodes, nil
}
