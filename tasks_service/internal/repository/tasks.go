package tasks_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	"github.com/liriquew/tasks_service/internal/lib/config"
	"github.com/liriquew/tasks_service/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrTaskNotInGroup = errors.New("task not in group")
)

type TaskRepository struct {
	db *sqlx.DB
}

const listTasksBatchSize = 10

func NewTaskRepository(cfg config.StorageConfig) (*TaskRepository, error) {
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
	return &TaskRepository{
		db: db,
	}, nil
}

func (s *TaskRepository) Close() error {
	return s.db.Close()
}

func (s *TaskRepository) SaveTask(ctx context.Context, task *tsks_pb.Task) (int64, error) {
	txn, err := s.db.Begin()
	if err != nil {
		return 0, err
	}
	defer func() {
		txn.Rollback()
	}()

	query := `INSERT INTO tasks (created_by, title, description, planned_time, actual_time)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err = txn.QueryRowContext(
		ctx,
		query,
		task.CreatedBy, task.Title, task.Description, task.PlannedTime, task.ActualTime,
	).Scan(&task.ID)
	if err != nil {
		return 0, err
	}

	if task.GroupID == 0 {
		err = txn.Commit()
		return task.ID, err
	}

	query = "INSERT INTO tasks_groups (task_id, group_id, assigned_to) VALUES ($1, $2, $3)"
	if _, err := txn.Exec(query, task.ID, task.GroupID, task.AssignedTo); err != nil {
		return 0, err
	}

	if err := txn.Commit(); err != nil {
		return 0, err
	}
	return task.ID, nil
}

func (s *TaskRepository) GetGroupTasks(ctx context.Context, groupID int64) ([]*models.Task, error) {
	query := `
	SELECT t.id, t.created_by, t.title, t.description, t.planned_time, t.actual_time, t.created_at, 
		tg.group_id, tg.assigned_to
	FROM tasks t JOIN tasks_groups tg ON t.id = tg.task_id WHERE tg.group_id=$1`

	var tasks []*models.Task
	if err := s.db.SelectContext(ctx, &tasks, query, groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return tasks, nil
}

func (s *TaskRepository) GetTaskByID(ctx context.Context, taskID int64) (*models.Task, error) {
	query := `
	SELECT 
		t.id, t.created_by, t.title, t.description, t.planned_time, t.actual_time, t.created_at, 
		tg.group_id, tg.assigned_to
	FROM tasks t LEFT JOIN tasks_groups tg ON t.id = tg.task_id WHERE t.id=$1
	`

	var task models.Task
	err := s.db.GetContext(ctx, &task, query, taskID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &task, nil
}

func (s *TaskRepository) GetTaskList(ctx context.Context, userID, padding int64) ([]*models.Task, error) {
	query := `
	SELECT 
		t.id, t.created_by, t.title, t.description, t.planned_time, t.actual_time, t.created_at, 
		tg.group_id, tg.assigned_to
	FROM tasks t LEFT JOIN tasks_groups tg ON t.id = tg.task_id WHERE t.created_by=$1 OR tg.assigned_to=$1 
	ORDER BY t.created_at DESC
	OFFSET $2 LIMIT $3
	`
	var tasks []*models.Task
	err := s.db.SelectContext(ctx, &tasks, query, userID, padding, listTasksBatchSize)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskRepository) UpdateTask(ctx context.Context, task *tsks_pb.Task) error {
	query := "UPDATE tasks SET %s WHERE created_by=$1 AND id=$2 RETURNING id"

	var fields []string
	args := []any{task.CreatedBy, task.ID}
	argPos := 2

	if task.ActualTime != 0 {
		argPos++
		fields = append(fields, fmt.Sprintf("actual_time=$%d", argPos))
		args = append(args, task.ActualTime)
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

	query = fmt.Sprintf(query, strings.Join(fields, ", "))

	if len(fields) != 0 {
		err := s.db.QueryRowContext(ctx, query, args...).Scan(&task.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
	}

	return nil
}

func (s *TaskRepository) UpdateGroupTask(ctx context.Context, task *tsks_pb.Task) error {
	query := "UPDATE tasks SET %s WHERE id=$1 RETURNING id"

	var fields []string
	args := []any{task.ID}
	argPos := 1

	if task.ActualTime != 0 {
		argPos++
		fields = append(fields, fmt.Sprintf("actual_time=$%d", argPos))
		args = append(args, task.ActualTime)
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

	query = fmt.Sprintf(query, strings.Join(fields, ", "))

	txn, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer txn.Rollback()

	if len(fields) != 0 {
		err = txn.QueryRowContext(ctx, query, args...).Scan(&task.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return ErrNotFound
			}
			return err
		}
	}

	if task.AssignedTo != 0 {
		query = "UPDATE tasks_groups SET assigned_to=$1 WHERE task_id=$2"
		rowsAffected, err := txn.ExecContext(ctx, query, task.AssignedTo, task.ID)
		if err != nil {
			return err
		}
		rows, err := rowsAffected.RowsAffected()
		if err != nil {
			return err
		}
		if rows == 0 {
			return ErrTaskNotInGroup
		}
	}

	if err := txn.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *TaskRepository) DeleteUserTask(ctx context.Context, userID, taskID int64) error {
	query := `
	DELETE FROM tasks 
	WHERE id=$1 AND created_by=$2`

	_, err := s.db.ExecContext(ctx, query, taskID, userID)
	if err != nil {
		return fmt.Errorf("error executing delete query: %w", err)
	}

	return nil
}

func (s *TaskRepository) DeleteGroupTask(ctx context.Context, taskID int64) error {
	query := `
	DELETE FROM tasks WHERE id=$1`

	_, err := s.db.ExecContext(ctx, query, taskID)
	if err != nil {
		return fmt.Errorf("error executing delete query: %w", err)
	}

	return nil
}

func (s *TaskRepository) TaskInAnyGroup(ctx context.Context, taskID int64) (int64, error) {
	query := "SELECT group_id FROM tasks_groups WHERE task_id=$2"

	var groupID int64
	if err := s.db.QueryRowContext(ctx, query, taskID).Scan(&groupID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrNotFound
		}
		return 0, err
	}

	return groupID, nil
}

func (s *TaskRepository) TaskInGroup(ctx context.Context, groupID, taskID int64) error {
	query := "SELECT 1 FROM tasks_groups WHERE group_id=$1 AND task_id=$2"

	var val int
	if err := s.db.QueryRowContext(ctx, query, groupID, taskID).Scan(&val); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

func (s *TaskRepository) GetTasks(ctx context.Context, tasksIDs []int64) ([]*models.Task, error) {
	query := `SELECT 
		t.id, t.created_by, t.title, t.description, t.planned_time, t.actual_time, t.created_at, 
		tg.group_id, tg.assigned_to
		FROM tasks t LEFT JOIN tasks_groups tg ON t.id = tg.task_id 
		WHERE id = ANY($1)`

	var tasks []*models.Task
	if err := s.db.SelectContext(ctx, &tasks, query, pq.Array(tasksIDs)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return tasks, nil
}

func (s *TaskRepository) Done(ctx context.Context, taskID int64, time float64) error {
	query := "UPDATE tasks SET actual_time=$1 WHERE id=$2"

	result, err := s.db.ExecContext(ctx, query, time, taskID)
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
