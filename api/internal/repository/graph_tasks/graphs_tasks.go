package graphstasks_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/liriquew/control_system/internal/models"
)

type GraphsTasksRepository struct {
	db *sqlx.DB
}

func NewGraphsTasksRepository(db *sqlx.DB) (*GraphsTasksRepository, error) {
	return &GraphsTasksRepository{
		db: db,
	}, nil
}

var (
	ErrNotFound = fmt.Errorf("not found")
)

func (r *GraphsTasksRepository) GetTasksFromNodes(ctx context.Context, nodes []*models.Node) ([]*models.Task, error) {
	query := `
		SELECT * FROM tasks WHERE id = ANY($1)
	`

	nodesIDs := make([]int64, 0, len(nodes))
	for _, node := range nodes {
		nodesIDs = append(nodesIDs, node.TaskID)
	}
	var tasks []*models.Task
	if err := r.db.SelectContext(ctx, &tasks, query, pq.Array(nodesIDs)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no rows in result query: %w", ErrNotFound)
		}

		return nil, err
	}

	return tasks, nil
}
