package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	predictions_client "time_manage/internal/grpc/client"
	"time_manage/internal/models"
	repository "time_manage/internal/repository/tasks"

	"github.com/jmoiron/sqlx"
)

type taskRepository interface {
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	GetTaskByID(ctx context.Context, uid, taskID int64) (*models.Task, error)
	GetTasksByUID(ctx context.Context, uid, limit, offset int64) ([]*models.Task, error)
	BeginTx(ctx context.Context) (*sqlx.Tx, error)
	UpdateTask(ctx context.Context, txn *sqlx.Tx, task *models.Task) error
	MarkModel(ctx context.Context, txn *sqlx.Tx, userID int64, modelFlag bool) (bool, error)
	DeleteTask(ctx context.Context, txn *sqlx.Tx, uid, taskID int64) error
}

type TaskService struct {
	taskRepo   taskRepository
	taskClient *predictions_client.Client
	infoLog    *log.Logger
	errorLog   *log.Logger
}

func NewTaskService(taskRepo taskRepository, taskClient *predictions_client.Client, infolog *log.Logger, errorLog *log.Logger) (*TaskService, error) {
	return &TaskService{
		taskRepo:   taskRepo,
		taskClient: taskClient,
		infoLog:    infolog,
		errorLog:   errorLog,
	}, nil
}

var (
	ErrNotFound = errors.New("not found")

	ErrInvalidTaskID = fmt.Errorf("invalid task id")
	ErrTaskNotFound  = fmt.Errorf("task not found")
	ErrInvalidParams = fmt.Errorf("ivalid argument")

	ErrNoCompletedTasks = fmt.Errorf("no completed tasks")
	ErrBadParams        = fmt.Errorf("bad params")

	ErrModelNotFound = fmt.Errorf("model not found")

	ErrBadTaskTile        = fmt.Errorf("bad task title")
	ErrBadTaskDescription = fmt.Errorf("bad task title")

	ErrNothingToUpdate = fmt.Errorf("there is nothing to update")
)

const LimitMaxValue int64 = 10

func (ts *TaskService) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	if task.Title == "" {
		return nil, ErrBadTaskTile
	}
	if task.Description == "" {
		return nil, ErrBadTaskDescription
	}

	task, err := ts.taskRepo.CreateTask(ctx, task)
	if err != nil {
		return nil, err
	}

	return task, nil
}

func (ts *TaskService) GetTaskByID(ctx context.Context, userID, taskID int64) (*models.Task, error) {
	if taskID <= 0 {
		return nil, ErrInvalidTaskID
	}

	task, err := ts.taskRepo.GetTaskByID(ctx, userID, taskID)
	if err != nil {
		// if errors.Is(err, repository.ErrInvalidTaskID) {
		// 	return nil, ErrInvalidTaskID
		// }
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}

	return task, nil
}

func (ts *TaskService) GetTaskListWithOffset(ctx context.Context, userID, limit, offset int64) ([]*models.Task, error) {
	if limit < 0 || offset < 0 {
		return nil, ErrInvalidParams
	}

	limit = min(limit, LimitMaxValue)

	tasks, err := ts.taskRepo.GetTasksByUID(ctx, userID, limit, offset)
	if err != nil {
		// TODO: check err
		return nil, err
	}

	return tasks, nil
}

func (ts *TaskService) UpdateTaskAndMarkModel(ctx context.Context, task *models.Task) (*models.Task, error) {
	// открывает транзакцию, в рамках нее обновляет задачу и помечает модель как не активную
	// в последствии при вызове predict модель надо будет заново обучить (вычислить) на gRPC-сервере
	const op = "UpdateTaskAndMarkModel"

	if task.Title == "" && task.Description == "" && task.PlannedTime == 0.0 && task.ActualTime == nil {
		return nil, ErrNothingToUpdate
	}

	txn, err := ts.taskRepo.BeginTx(ctx)
	if err != nil {
		ts.errorLog.Println(op, err)
		return nil, err
	}
	defer func() {
		err = txn.Rollback()
	}()

	err = ts.taskRepo.UpdateTask(ctx, txn, task)
	if err != nil {
		// if errors.Is(err, repository.ErrInvalidTaskID) {
		// 	return nil, ErrInvalidTaskID
		// }
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, ErrTaskNotFound
		}

		return nil, err
	}

	_, err = ts.taskRepo.MarkModel(ctx, txn, task.UserID, false)
	if err != nil {
		if errors.Is(err, repository.ErrModelNotFound) {
			return nil, ErrModelNotFound
		}
		return nil, err
	}

	if err := txn.Commit(); err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	return task, nil
}

func (ts *TaskService) DeleteTaskAndMarkModel(ctx context.Context, userID, taskID int64) (bool, error) {
	// открывает транзакцию, в рамках нее обновляет задачу и помечает модель как не активную
	// в последствии при вызове predict модель надо будет заново обучить (вычислить) на gRPC-сервере
	const op = "DeleteTaskAndMarkModel"

	txn, err := ts.taskRepo.BeginTx(ctx)
	if err != nil {
		ts.errorLog.Println(op, err)
		return false, err
	}
	defer func() {
		err = txn.Rollback()
	}()

	err = ts.taskRepo.DeleteTask(ctx, txn, userID, taskID)
	if err != nil {
		return false, err
	}

	_, err = ts.taskRepo.MarkModel(ctx, txn, userID, false)
	if err != nil {
		if errors.Is(err, repository.ErrModelNotFound) {
			return false, ErrModelNotFound
		}
		return false, err
	}

	if err := txn.Commit(); err != nil {
		return false, fmt.Errorf("commit failed: %w", err)
	}

	return true, nil
}

func (ts *TaskService) PredictTaskCompletionTime(ctx context.Context, userID int64, plannedTime float64) (float64, error) {
	actualTime, err := ts.taskClient.Predict(ctx, userID, plannedTime)
	if err != nil {
		if errors.Is(err, predictions_client.ErrFailedPrecondition) {
			return 0, ErrNoCompletedTasks
		} else if errors.Is(err, predictions_client.ErrInvalidArgument) {
			return 0, ErrBadParams
		}
		ts.errorLog.Println("gRPC error:", err)
		return 0, err
	}

	return actualTime, nil
}
