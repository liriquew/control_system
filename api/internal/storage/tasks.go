package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var (
	ErrInvalidTaskID = fmt.Errorf("Invalid TaskID <= 0")
	ErrTaskNotFound  = fmt.Errorf("Task not found")
)

type Task struct {
	ID          int64     `json:"id" db:"id"`
	UserID      int64     `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	PlannedTime float64   `json:"planned_time" db:"planned_time"` // планируемое время выполнения
	ActualTime  *float64  `json:"actual_time" db:"actual_time"`   // фактическое время выполнения
	CreatedAt   time.Time `json:"created_at" db:"created_at"`     // Время создания задачи
}

func (s *Storage) CreateTask(ctx context.Context, uid int64, task *Task) (int64, error) {
	query := `INSERT INTO tasks (user_id, title, description, planned_time, actual_time)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	fmt.Println(uid, task)
	err := s.db.QueryRowContext(
		ctx,
		query,
		uid, task.Title, task.Description, task.PlannedTime, nil,
	).Scan(&task.ID)

	if err != nil {
		return 0, err
	}

	return task.ID, err
}

func (s *Storage) GetTaskByID(ctx context.Context, uid, taskID int64) (*Task, error) {
	if taskID <= 0 {
		return nil, ErrInvalidTaskID
	}

	query := "SELECT * FROM tasks WHERE user_id=$1 AND id=$2"

	var task Task
	err := s.db.GetContext(ctx, &task, query, uid, taskID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	return &task, nil
}

func (s *Storage) UpdateTask(ctx context.Context, uid, taskID int64, task *Task) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	query := "UPDATE tasks SET title=$1, description=$2, planned_time=$3, actual_time=$4 WHERE user_id=$5 AND id=$6"

	result, err := s.db.ExecContext(
		ctx,
		query,
		task.Title, task.Description, task.PlannedTime, task.ActualTime, uid, taskID,
	)

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

	return err
}

func (s *Storage) DeleteTask(ctx context.Context, uid, taskID int64) error {
	if taskID <= 0 {
		return ErrInvalidTaskID
	}

	query := "DELETE FROM tasks WHERE user_id=$1 AND id=$2"

	_, err := s.db.ExecContext(
		ctx,
		query,
		uid, taskID,
	)

	if err != nil {
		return fmt.Errorf("error executing delete query: %w", err)
	}

	return nil
}
