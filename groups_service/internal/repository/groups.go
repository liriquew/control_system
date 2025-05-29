package groups_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	grpc_pb "github.com/liriquew/control_system/services_protos/groups_service"
	"github.com/liriquew/groups_service/internal/lib/config"
	"github.com/liriquew/groups_service/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type GroupsRepository struct {
	db *sqlx.DB
}

const listTasksBatchSize = 10

func NewGroupRepository(cfg config.StorageConfig) (*GroupsRepository, error) {
	const op = "storage.postgres.New"

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = db.Ping(); err != nil {
		panic(op + ":" + err.Error())
	}

	return &GroupsRepository{
		db: db,
	}, nil
}

func (s *GroupsRepository) Close() error {
	return s.db.Close()
}

var (
	ErrDenied = errors.New("access denied")

	ErrNotFound        = errors.New("not found")
	ErrNotExists       = errors.New("not exists")
	ErrNothingToUpdate = errors.New("nothing to update (empty updatable fields)")
	ErrInvalideRole    = errors.New("invalide role param")
	ErrWrongOwnerID    = errors.New("user with owner id not found")
	ErrAlreadyInGroup  = errors.New("already in group")
)

func (gr *GroupsRepository) CheckAccess(ctx context.Context, userID, groupID int64) error {
	query := `
		SELECT 1 FROM group_members gm 
		WHERE gm.user_id = $1 AND gm.group_id = $2
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
		WHERE gm.user_id = $1 AND gm.group_id = $2 AND gm.role = ANY($3)
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

func (gr *GroupsRepository) CreateGroup(ctx context.Context, group *grpc_pb.Group) (int64, error) {
	txn, err := gr.db.Begin()
	if err != nil {
		return 0, err
	}
	defer txn.Rollback()
	query := "INSERT INTO groups (owner_id, name, description) VALUES ($1, $2, $3) RETURNING id"

	err = txn.QueryRowContext(ctx, query, group.OwnerID, group.Name, group.Description).Scan(&group.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			return 0, fmt.Errorf("database error: %s", pqErr.Message)
		}
		return 0, err
	}

	query = "INSERT INTO group_members (user_id, group_id, role) VALUES ($1, $2, $3)"
	if _, err := txn.ExecContext(ctx, query, group.OwnerID, group.ID, "admin"); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23514": // Код ошибки для CHECK violation
				if pqErr.Constraint == "no_group_member_role" {
					return 0, ErrInvalideRole
				}
			default:
				return 0, fmt.Errorf("database error: %s", pqErr.Message)
			}
		}

		return 0, err
	}

	if err := txn.Commit(); err != nil {
		return 0, err
	}

	return group.ID, nil
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

func (gr *GroupsRepository) ListUserGroups(ctx context.Context, userID int64, offset int64) ([]*models.Group, error) {
	query := `
		SELECT * FROM groups
		WHERE id IN (SELECT group_id FROM group_members WHERE user_id=$1) ORDER BY created_at OFFSET $2 LIMIT 10
	`

	var groups []*models.Group
	err := gr.db.SelectContext(ctx, &groups, query, userID, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("groups not found%w", ErrNotFound)
		}
		return nil, err
	}

	return groups, err
}

func (gr *GroupsRepository) DeleteGroup(ctx context.Context, ownerID, groupID int64) error {
	query := `DELETE FROM groups WHERE id=$2 AND owner_id=$1`

	result, err := gr.db.ExecContext(ctx, query, ownerID, groupID)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (gr *GroupsRepository) UpdateGroup(ctx context.Context, group *grpc_pb.Group) error {
	var fields []string
	args := []any{
		0: group.ID,
	}
	argPos := 2

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

	query := fmt.Sprintf("UPDATE groups SET %s WHERE id=$1", strings.Join(fields, ", "))

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

func (gr *GroupsRepository) ListGroupMembers(ctx context.Context, groupID int64) ([]*models.GroupMember, error) {
	query := `
        SELECT user_id, role FROM group_members
        WHERE group_id = $1
	`

	var users []*models.GroupMember
	err := gr.db.SelectContext(ctx, &users, query, groupID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return users, nil
}

func (gr *GroupsRepository) AddGroupMember(ctx context.Context, member *grpc_pb.GroupMember) error {
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
		return ErrInvalideRole
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
					return fmt.Errorf("group with id %d does not exist%w", member.GroupID, ErrNotExists)
				}
			case "23505": // Код ошибки для FOREIGN KEY violation
				switch pqErr.Constraint {
				case "group_members_pkey":
					return fmt.Errorf("user already in group%w", ErrAlreadyInGroup)
				}
			case "23514": // Код ошибки для CHECK violation
				if pqErr.Constraint == "no_group_member_role" {
					return ErrInvalideRole
				}
			default:
				return fmt.Errorf("database error: %s", pqErr.Message)
			}
		}
		return err
	}

	return nil
}

func (gr *GroupsRepository) RemoveGroupMember(ctx context.Context, member *grpc_pb.GroupMember) error {
	query := `
		DELETE FROM group_members 
		WHERE 
			group_id=$1 AND 
			user_id=$2
	`

	result, err := gr.db.ExecContext(ctx, query, member.GroupID, member.UserID)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (gr *GroupsRepository) ChangeMemberRole(ctx context.Context, ownerID int64, member *grpc_pb.GroupMember) error {
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
        UPDATE group_members SET role=$1 WHERE group_id=$2 AND user_id=$3
    `

	result, err := gr.db.ExecContext(ctx, query, member.Role, member.GroupID, member.UserID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23514": // Код ошибки для CHECK violation
				if pqErr.Constraint == "no_group_member_role" {
					return ErrInvalideRole
				}
			default:
				return fmt.Errorf("database error: %s", pqErr.Message)
			}
		}

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
