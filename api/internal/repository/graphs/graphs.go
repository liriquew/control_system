package graphs_repository

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

type GraphsRepository struct {
	db *sqlx.DB
}

func NewGraphsRepository(db *sqlx.DB) (*GraphsRepository, error) {
	return &GraphsRepository{
		db: db,
	}, nil
}

var (
	ErrDenied = errors.New("access denied")

	ErrNotFound               = errors.New("not found")
	ErrNotExists              = errors.New("")
	ErrNodeNotFound           = errors.New("node not found")
	ErrDependencyNotFound     = errors.New("dependency not found")
	ErrSelfDependencyRejected = fmt.Errorf("self dependensy is rejected")
	ErrNothingToUpdate        = errors.New("")
)

func (gr *GraphsRepository) CheckEditorPermission(ctx context.Context, userID, graphID int64) error {
	query := `
		SELECT g.id FROM graphs g 
		JOIN groups grp ON g.group_id = grp.id
		JOIN group_members gm ON gm.group_id = grp.id
		WHERE gm.user_id=$1 AND g.id=$2 AND gm.role IN ('admin', 'editor')
	`

	var id int64
	if err := gr.db.QueryRowContext(ctx, query, userID, graphID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrDenied
		}
		return err
	}

	return nil
}

func (gr *GraphsRepository) CheckUserPermission(ctx context.Context, userID, graphID int64) error {
	query := `
		SELECT g.id FROM graphs g 
		JOIN groups grp ON g.group_id = grp.id
		JOIN group_members gm ON gm.group_id = grp.id
		WHERE gm.user_id=$1 AND g.id=$2
	`

	var id int64
	if err := gr.db.QueryRowContext(ctx, query, userID, graphID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrDenied
		}
		return err
	}
	return nil
}

func (gr *GraphsRepository) GetGraph(ctx context.Context, graphID int64) (*entities.GraphWithNodes, error) {
	query := "SELECT * FROM graphs WHERE id=$1"

	var graph models.Graph
	if err := gr.db.GetContext(ctx, &graph, query, graphID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	query = "SELECT * FROM nodes WHERE graph_id=$1"

	var nodes []*models.Node
	if err := gr.db.SelectContext(ctx, &nodes, query, graphID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to get nodes: %w", err)
		}
		// Если узлов нет, оставляем nodes пустым
	}

	query = "SELECT to_node_id FROM dependencies WHERE graph_id=$1 AND from_node_id=$2"

	for _, node := range nodes {
		if err := gr.db.SelectContext(ctx, &node.DependencyNodeIDs, query, graphID, node.ID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("failed to get dependencies: %w", err)
			}
			// Если зависимостей нет, оставляем deps пустым
		}
	}

	return &entities.GraphWithNodes{
		GraphInfo: graph,
		Nodes:     nodes,
	}, nil
}

func (gr *GraphsRepository) CreateNode(ctx context.Context, userID, graphID int64, node *models.Node) (*models.Node, error) {
	txn, err := gr.db.Begin()
	if err != nil {
		return nil, err
	}
	defer txn.Rollback()

	query := "INSERT INTO nodes (graph_id%s) VALUES ($1%s) RETURNING id"
	args := []any{graphID}
	var fields, placeholders []string
	argPos := 2
	if node.TaskID != 0 {
		args = append(args, node.TaskID)
		fields = append(fields, "task_id")
		placeholders = append(placeholders, fmt.Sprintf("$%d", argPos))
		argPos++
	}

	if node.TaskID != 0 && node.AssignedTo != nil && *node.AssignedTo != 0 {
		args = append(args, *node.AssignedTo)
		fields = append(fields, "assigned_to")
		placeholders = append(placeholders, fmt.Sprintf("$%d", argPos))
		argPos++
	}

	if len(fields) != 0 {
		query = fmt.Sprintf(query, ", "+strings.Join(fields, ", "), ", "+strings.Join(placeholders, ", "))
	} else {
		query = fmt.Sprintf(query, "", "")
	}
	if err := txn.QueryRowContext(ctx, query, args...).Scan(&node.ID); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // Код ошибки для FOREIGN KEY violation
				switch pqErr.Constraint {
				case "fk_graph_dependency":
					return nil, fmt.Errorf("graph with id %d does not exist%w", graphID, ErrNotExists)
				case "fk_task_node":
					return nil, fmt.Errorf("task with id %d does not exist%w", node.TaskID, ErrNotExists)
				}
			default:
				return nil, fmt.Errorf("database error: %s", pqErr.Message)
			}
		}
		return nil, err
	}

	query = "INSERT INTO dependencies (graph_id, from_node_id, to_node_id) VALUES ($1, $2, $3)"

	nodeIDs := make(map[int64]any, 0)
	for _, depNodeID := range node.DependencyNodeIDs {
		if _, ok := nodeIDs[depNodeID]; ok {
			continue
		}
		nodeIDs[depNodeID] = struct{}{}

		if _, err := txn.ExecContext(ctx, query, graphID, node.ID, depNodeID); err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23503": // Код ошибки для FOREIGN KEY violation
					switch pqErr.Constraint {
					case "fk_graph_dependency":
						return nil, fmt.Errorf("graph with id %d does not exist%w", graphID, ErrNotExists)
					case "fk_from_node_dependency":
						return nil, fmt.Errorf("node with id %d does not exist%w", depNodeID, ErrNotExists)
					case "fk_to_node_dependency":
						return nil, fmt.Errorf("node with id %d does not exist%w", node.ID, ErrNotExists)
					}
				case "23514": // Код ошибки для CHECK violation
					if pqErr.Constraint == "no_self_dependency" {
						return nil, ErrSelfDependencyRejected
					}
				}
			}
			return nil, err
		}
	}

	if err := txn.Commit(); err != nil {
		return nil, err
	}

	return node, nil
}

func (gr *GraphsRepository) GetNode(ctx context.Context, graphID, nodeID int64) (*models.Node, error) {
	query := `
		SELECT * FROM nodes WHERE id=$1
	`

	var node models.Node
	if err := gr.db.GetContext(ctx, &node, query, nodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNodeNotFound
		}
		return nil, fmt.Errorf("repository: error while getting node: %w", err)
	}

	query = `
		SELECT to_node_id FROM dependencies WHERE graph_id=$1 AND from_node_id=$2
	`

	if err := gr.db.SelectContext(ctx, &node.DependencyNodeIDs, query, graphID, nodeID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("repository: error while getting dependencies: %w", err)
		}
		// Если зависимостей нет, оставляем deps пустым
	}

	return &node, nil
}

func (gr *GraphsRepository) UpdateNode(ctx context.Context, userID, graphID int64, node *models.Node) error {
	query := "UPDATE nodes SET %s WHERE id=$%d"

	var placeholders []string
	var args []any
	argPos := 1
	if node.TaskID != 0 {
		placeholders = append(placeholders, fmt.Sprintf("task_id=$%d", argPos))
		args = append(args, node.TaskID)
		argPos++
	}
	// TODO: pls add some trigger func in postgres to check if task already exists
	// u cant add worker to task without add task btw
	// created in future func must throw exeption idk and this exception must be
	// checked in golang right here :)
	if node.AssignedTo != nil && *node.AssignedTo != 0 {
		placeholders = append(placeholders, fmt.Sprintf("task_id=$%d", argPos))
		args = append(args, node.TaskID)
		argPos++
	}

	if len(placeholders) == 0 {
		return ErrNothingToUpdate
	}

	query = fmt.Sprintf(query, strings.Join(placeholders, ", "), argPos)
	args = append(args, node.ID)

	result, err := gr.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		return err
	} else if rowsAffected == 0 {
		return ErrNotExists
	}

	return nil
}

func (gr *GraphsRepository) RemoveNode(ctx context.Context, userID, graphID, nodeID int64) error {
	query := "DELETE FROM nodes WHERE id=$1"

	result, err := gr.db.ExecContext(ctx, query, nodeID)
	if err != nil {
		return nil
	}

	if rowsAffected, err := result.RowsAffected(); rowsAffected == 0 {
		return ErrNotExists
	} else if err != nil {
		return nil
	}

	return nil
}

func (gr *GraphsRepository) GetDependencies(ctx context.Context, userID, graphID, nodeID int64) (*models.Node, error) {
	nodeWithDeps := models.Node{}
	query := "SELECT * FROM nodes WHERE id=$1"

	if err := gr.db.GetContext(ctx, &nodeWithDeps, query, nodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNodeNotFound
		}
		return nil, err
	}

	query = "SELECT to_node_id FROM dependencies WHERE graph_id=$1 AND from_node_id=$2"

	if err := gr.db.SelectContext(ctx, &nodeWithDeps.DependencyNodeIDs, query, graphID, nodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDependencyNotFound
		}
		return nil, err
	}

	return &nodeWithDeps, nil
}

func (gr *GraphsRepository) AddDependency(ctx context.Context, userID, graphID int64, dependency *models.Dependency) (*models.Dependency, error) {
	query := "INSERT INTO dependencies (from_node_id, to_node_id, graph_id) VALUES ($1, $2, $3)"
	fmt.Println(dependency.FromNodeID, dependency.ToNodeID, graphID)
	if _, err := gr.db.ExecContext(ctx, query, dependency.FromNodeID, dependency.ToNodeID, graphID); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // Код ошибки для FOREIGN KEY violation
				switch pqErr.Constraint {
				case "fk_graph_dependency":
					return nil, fmt.Errorf("graph with id %d does not exist%w", graphID, ErrNotExists)
				case "fk_from_node_dependency":
					return nil, fmt.Errorf("node with id %d does not exist%w", dependency.FromNodeID, ErrNotExists)
				case "fk_to_node_dependency":
					return nil, fmt.Errorf("node with id %d does not exist%w", dependency.ToNodeID, ErrNotExists)
				}
			case "23514": // Код ошибки для CHECK violation
				if pqErr.Constraint == "no_self_dependency" {
					return nil, ErrSelfDependencyRejected
				}
			}
		}
		return nil, err
	}

	return dependency, nil
}

func (gr *GraphsRepository) RemoveDependensy(ctx context.Context, userID, graphID int64, dependency *models.Dependency) error {
	query := "DELETE FROM dependencies WHERE from_node_id=$1 AND to_node_id=$2"

	result, err := gr.db.ExecContext(ctx, query, dependency.FromNodeID, dependency.ToNodeID)
	if err != nil {
		return err
	}

	if rowsAffected, err := result.RowsAffected(); rowsAffected == 0 {
		return ErrNotExists
	} else if err != nil {
		return nil
	}

	return nil
}
