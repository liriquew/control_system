package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/liriquew/graphs_service/internal/entities"
	"github.com/liriquew/graphs_service/internal/lib/config"
	"github.com/liriquew/graphs_service/internal/models"

	grph_pb "github.com/liriquew/control_system/services_protos/graphs_service"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type GraphsRepository struct {
	db *sqlx.DB
}

const listTasksBatchSize = 10

func NewGraphsRepository(cfg config.StorageConfig) (*GraphsRepository, error) {
	const op = "storage.postgres.New"

	connStr := fmt.Sprintf("postgres://%s:%s@localhost:%d/%s?sslmode=disable",
		cfg.Username,
		cfg.Password,
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

	fmt.Println("DB CONNECT OK")
	return &GraphsRepository{
		db: db,
	}, nil
}

func (s *GraphsRepository) Close() error {
	return s.db.Close()
}

var (
	ErrNotFound  = errors.New("not found")
	ErrNotExists = errors.New("not exists")

	ErrSelfDependencyRejected = errors.New("found self dependency")
	ErrNothingToUpdate        = errors.New("nothing to update")
	ErrNodeNotFound           = errors.New("node not found")
	ErrDependencyNotFound     = errors.New("dependency not found")
	ErrTaskAlreadyInNode      = errors.New("task already in node")
)

func (r *GraphsRepository) GetGraphGroup(ctx context.Context, graphID int64) (int64, error) {
	query := `SELECT group_id FROM graphs WHERE id=$1`

	var groupID int64
	if err := r.db.QueryRowContext(ctx, query, graphID).Scan(&groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}

		return 0, err
	}

	return groupID, nil
}

func (r *GraphsRepository) CreateGraph(ctx context.Context, graph *grph_pb.Graph, nodes []*grph_pb.Node) (int64, error) {
	txn, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer txn.Rollback()

	// TODO: add constraint in postgres unique name and check err
	graphQuery := `INSERT INTO graphs (group_id, created_by, name) VALUES ($1, $2, $3) RETURNING id`
	err = txn.QueryRowContext(ctx, graphQuery, graph.GroupID, graph.CreatedBy, graph.Name).Scan(&graph.ID)
	if err != nil {
		return 0, fmt.Errorf("error while creating graph: %w", err)
	}

	// Узлы
	nodeQuery := `INSERT INTO nodes (graph_id, task_id) VALUES ($1, $2) RETURNING id`
	nodeIDMap := make(map[int64]int64, len(nodes))
	for _, node := range nodes {
		var nodeID int64
		err = txn.QueryRowContext(ctx, nodeQuery, graph.ID, node.TaskID).Scan(&nodeID)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23505":
					if pqErr.Constraint == "uq_task_id" {
						return 0, fmt.Errorf("task with id %d alredy in some node: %w", node.TaskID, ErrTaskAlreadyInNode)
					}
				default:
					return 0, fmt.Errorf("database error: %s", pqErr.Message)
				}
			}
			return 0, fmt.Errorf("error while creating node: %w", err)
		}
		nodeIDMap[node.ID] = nodeID
		node.ID = nodeID
	}

	// Зависимости
	// TODO: add constraint in postgres nodes in one graph
	dependencyQuery := `INSERT INTO dependencies (from_node_id, to_node_id, graph_id) VALUES ($1, $2, $3)`
	stmt, err := txn.PrepareContext(ctx, dependencyQuery)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	for _, node := range nodes {
		fromNodeID := node.ID
		for _, dependencyNodeDummyID := range node.DependencyNodeIDs {
			toNodeID := nodeIDMap[dependencyNodeDummyID]
			_, err = stmt.ExecContext(ctx, fromNodeID, toNodeID, graph.ID)
			if err != nil {
				if pqErr, ok := err.(*pq.Error); ok {
					switch pqErr.Code {
					case "23514": // Код ошибки для CHECK violation
						if pqErr.Constraint == "no_self_dependency" {
							return 0, ErrSelfDependencyRejected
						}
					default:
						return 0, fmt.Errorf("database error: %s", pqErr.Message)
					}
				}
				return 0, fmt.Errorf("error while creating dependency: %w", err)
			}
			node.DependencyNodeIDs = append(node.DependencyNodeIDs, toNodeID)
		}
	}

	if err = txn.Commit(); err != nil {
		return 0, err
	}

	return graph.ID, nil
}

func (r *GraphsRepository) ListGroupGraphs(ctx context.Context, userID int64, groupID int64) ([]*entities.GraphWithNodes, error) {
	query := `SELECT * FROM graphs WHERE group_id=$1`

	var graphs []models.Graph
	err := r.db.SelectContext(ctx, &graphs, query, groupID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("error while getting graphs: %w", err)
	}

	graphsWithNodes := make([]*entities.GraphWithNodes, len(graphs))
	for i := range graphs {
		graphsWithNodes[i] = &entities.GraphWithNodes{
			GraphInfo: &graphs[i],
		}
	}

	queryNodes := "SELECT * FROM nodes WHERE graph_id=$1"
	queryDependencies := "SELECT to_node_id FROM dependencies WHERE from_node_id=$1"
	for i, graph := range graphs {
		if err := r.db.SelectContext(ctx, &graphsWithNodes[i].Nodes, queryNodes, graph.ID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}

			return nil, fmt.Errorf("error while getting nodes: %w", err)
		}

		for _, node := range graphsWithNodes[i].Nodes {
			if err := r.db.SelectContext(ctx, &node.DependencyNodeIDs, queryDependencies, node.ID); err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					continue
				}

				return nil, fmt.Errorf("error while getting dependencies: %w", err)
			}
		}
	}

	return graphsWithNodes, nil
}

func (r *GraphsRepository) GetGraph(ctx context.Context, graphID int64) (*entities.GraphWithNodes, error) {
	query := "SELECT * FROM graphs WHERE id=$1"

	var graph models.Graph
	if err := r.db.GetContext(ctx, &graph, query, graphID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get graph: %w", err)
	}

	query = "SELECT * FROM nodes WHERE graph_id=$1"

	var nodes []*models.Node
	if err := r.db.SelectContext(ctx, &nodes, query, graphID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("failed to get nodes: %w", err)
		}
		// Если узлов нет, оставляем nodes пустым
	}

	query = "SELECT to_node_id FROM dependencies WHERE graph_id=$1 AND from_node_id=$2"

	for _, node := range nodes {
		if err := r.db.SelectContext(ctx, &node.DependencyNodeIDs, query, graphID, node.ID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, fmt.Errorf("failed to get dependencies: %w", err)
			}
			// Если зависимостей нет, оставляем deps пустым
		}
	}

	return &entities.GraphWithNodes{
		GraphInfo: &graph,
		Nodes:     nodes,
	}, nil
}

func (r *GraphsRepository) CreateNode(ctx context.Context, node *grph_pb.Node) (int64, error) {
	txn, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	defer txn.Rollback()

	query := "INSERT INTO nodes (graph_id, task_id) VALUES ($1, $2) RETURNING id"
	if err := txn.QueryRowContext(ctx, query, node.GraphID, node.TaskID).Scan(&node.ID); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // Код ошибки для FOREIGN KEY violation
				switch pqErr.Constraint {
				case "fk_graph_dependency":
					return 0, fmt.Errorf("graph with id %d does not exist %w", node.GraphID, ErrNotExists)
				}
			case "23505":
				if pqErr.Constraint == "uq_task_id" {
					return 0, fmt.Errorf("task with id %d alredy in some node: %w", node.TaskID, ErrTaskAlreadyInNode)
				}
			default:
				return 0, fmt.Errorf("database error: %s", pqErr.Message)
			}
		}
		return 0, err
	}

	query = "INSERT INTO dependencies (graph_id, from_node_id, to_node_id) VALUES ($1, $2, $3)"
	// TOOD: check is nodes in one graph (write postgres constraint)
	nodeIDs := make(map[int64]any, 0)
	for _, depNodeID := range node.DependencyNodeIDs {
		if _, ok := nodeIDs[depNodeID]; ok {
			continue
		}
		nodeIDs[depNodeID] = struct{}{}
		if _, err := txn.ExecContext(ctx, query, node.GraphID, node.ID, depNodeID); err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				switch pqErr.Code {
				case "23503": // Код ошибки для FOREIGN KEY violation
					switch pqErr.Constraint {
					case "fk_graph_dependency":
						return 0, fmt.Errorf("graph with id %d does not exist%w", node.GraphID, ErrNotExists)
					case "fk_from_node_dependency":
						return 0, fmt.Errorf("node with id %d does not exist%w", node.ID, ErrNotExists)
					case "fk_to_node_dependency":
						return 0, fmt.Errorf("node with id %d does not exist%w", depNodeID, ErrNotExists)
					}
				case "23514": // Код ошибки для CHECK violation
					if pqErr.Constraint == "no_self_dependency" {
						return 0, ErrSelfDependencyRejected
					}
				default:
					return 0, fmt.Errorf("database error: %s", pqErr.Message)
				}
			}
			return 0, err
		}
	}

	if err := txn.Commit(); err != nil {
		return 0, err
	}

	return node.ID, nil
}

func (r *GraphsRepository) GetNode(ctx context.Context, graphID, nodeID int64) (*models.Node, error) {
	query := `SELECT * FROM nodes WHERE id=$1`

	var node models.Node
	if err := r.db.GetContext(ctx, &node, query, nodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("repository: error while getting node: %w", err)
	}

	query = `SELECT to_node_id FROM dependencies WHERE graph_id=$1 AND from_node_id=$2`

	if err := r.db.SelectContext(ctx, &node.DependencyNodeIDs, query, graphID, nodeID); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("repository: error while getting dependencies: %w", err)
		}
		// Если зависимостей нет, оставляем deps пустым
	}

	return &node, nil
}

func (r *GraphsRepository) UpdateNode(ctx context.Context, node *grph_pb.Node) error {
	query := "UPDATE nodes SET task_id=$1 WHERE id=$2"

	result, err := r.db.ExecContext(ctx, query, node.TaskID, node.ID)
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

func (r *GraphsRepository) RemoveNode(ctx context.Context, nodeID int64) error {
	query := "DELETE FROM nodes WHERE id=$1"

	result, err := r.db.ExecContext(ctx, query, nodeID)
	if err != nil {
		return nil
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		return err
	} else if rowsAffected == 0 {
		return nil
	}

	return nil
}

func (r *GraphsRepository) GetDependencies(ctx context.Context, nodeID int64) (*models.Node, error) {
	nodeWithDeps := models.Node{}
	query := "SELECT * FROM nodes WHERE id=$1"
	if err := r.db.GetContext(ctx, &nodeWithDeps, query, nodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNodeNotFound
		}
		return nil, err
	}

	query = "SELECT to_node_id FROM dependencies WHERE from_node_id=$1"
	if err := r.db.SelectContext(ctx, &nodeWithDeps.DependencyNodeIDs, query, nodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDependencyNotFound
		}
		return nil, err
	}

	return &nodeWithDeps, nil
}

func (r *GraphsRepository) AddDependency(ctx context.Context, graphID int64, dependency *grph_pb.Dependency) error {
	query := "INSERT INTO dependencies (from_node_id, to_node_id, graph_id) VALUES ($1, $2, $3)"
	if _, err := r.db.ExecContext(ctx, query, dependency.FromNodeID, dependency.ToNodeID, graphID); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // Код ошибки для FOREIGN KEY violation
				switch pqErr.Constraint {
				case "fk_from_node_dependency":
					return fmt.Errorf("node with id %d does not exist%w", dependency.FromNodeID, ErrNotExists)
				case "fk_to_node_dependency":
					return fmt.Errorf("node with id %d does not exist%w", dependency.ToNodeID, ErrNotExists)
				}
			case "23514": // Код ошибки для CHECK violation
				if pqErr.Constraint == "no_self_dependency" {
					return ErrSelfDependencyRejected
				}
			}
		}
		return err
	}

	return nil
}

func (r *GraphsRepository) RemoveDependensy(ctx context.Context, dependency *grph_pb.Dependency) error {
	query := "DELETE FROM dependencies WHERE from_node_id=$1 AND to_node_id=$2"

	result, err := r.db.ExecContext(ctx, query, dependency.FromNodeID, dependency.ToNodeID)
	if err != nil {
		return err
	}

	if rowsAffected, err := result.RowsAffected(); err != nil {
		return err
	} else if rowsAffected == 0 {
		return err
	}

	return nil
}

func (r *GraphsRepository) TaskInNode(ctx context.Context, taskID int64) (int64, error) {
	query := "SELECT id FROM nodes WHERE task_id=$1"

	var nodeID int64
	if err := r.db.QueryRowContext(ctx, query, taskID).Scan(&nodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}

		return 0, err
	}

	return nodeID, nil
}
