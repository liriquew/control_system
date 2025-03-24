package tasks_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/liriquew/control_system/internal/models"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	ErrTaskNotFound  = fmt.Errorf("task not found")
	ErrModelNotFound = fmt.Errorf("user model not found")
	ErrNotFound      = fmt.Errorf("not found")
)

type TaskRepository struct {
	db *sqlx.DB
}

func NewTaskRepository(db *sqlx.DB) (*TaskRepository, error) {
	return &TaskRepository{
		db: db,
	}, nil
}

func (s *TaskRepository) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	query := `INSERT INTO tasks (user_id, title, description, planned_time, actual_time)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := s.db.QueryRowContext(
		ctx,
		query,
		task.UserID, task.Title, task.Description, task.PlannedTime, nil,
	).Scan(&task.ID)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (s *TaskRepository) GetTaskByID(ctx context.Context, uid, taskID int64) (*models.Task, error) {
	query := "SELECT * FROM tasks WHERE user_id=$1 AND id=$2"

	var task models.Task
	err := s.db.GetContext(ctx, &task, query, uid, taskID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	return &task, nil
}

func (s *TaskRepository) GetTasksByUID(ctx context.Context, userID, limit, offset int64) ([]*models.Task, error) {
	query := "SELECT * FROM TASKS WHERE user_id=$1 OFFSET $2 LIMIT $3"

	var tasks []*models.Task
	err := s.db.SelectContext(ctx, &tasks, query, userID, offset, limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return tasks, nil
}

func (s *TaskRepository) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return s.db.Beginx()
}

func (s *TaskRepository) UpdateTask(ctx context.Context, txn *sqlx.Tx, task *models.Task) error {
	var fields []string
	var args []interface{}
	argPos := 0
	if task.ActualTime != nil {
		argPos++
		fields = append(fields, fmt.Sprintf("actual_time=$%d", argPos))
		args = append(args, *task.ActualTime)
	}
	if task.PlannedTime != 0 {
		argPos++
		fields = append(fields, fmt.Sprintf("planned_time=$%d", argPos))
		args = append(args, task.PlannedTime)
	}
	if task.Description != "" {
		argPos++
		fields = append(fields, fmt.Sprintf("description=$%d", argPos))
		args = append(args, task.Description)
	}
	if task.Title != "" {
		argPos++
		fields = append(fields, fmt.Sprintf("title=$%d", argPos))
		args = append(args, task.Title)
	}

	args = append(args, task.UserID, task.ID)

	query := fmt.Sprintf("UPDATE tasks SET %s WHERE user_id=$%d AND id=$%d", strings.Join(fields, ", "), argPos+1, argPos+2)

	result, err := txn.ExecContext(ctx, query, args...)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrTaskNotFound
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTaskNotFound
	}

	return nil
}

func (s *TaskRepository) MarkModel(ctx context.Context, txn *sqlx.Tx, userID int64, markFlag bool) (bool, error) {
	query := "UPDATE models SET is_active=$1"

	result, err := txn.ExecContext(ctx, query, markFlag)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrModelNotFound
		}
		return false, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowsAffected == 0 {
		return false, nil
	}

	return true, nil
}

func (s *TaskRepository) DeleteTask(ctx context.Context, txn *sqlx.Tx, userID, taskID int64) error {
	query := "DELETE FROM tasks WHERE user_id=$1 AND id=$2"

	_, err := txn.ExecContext(
		ctx,
		query,
		userID, taskID,
	)

	if err != nil {
		return fmt.Errorf("error executing delete query: %w", err)
	}

	return nil
}
