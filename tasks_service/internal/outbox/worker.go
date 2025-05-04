package outbox

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/liriquew/tasks_service/internal/models"
	"github.com/liriquew/tasks_service/pkg/logger/sl"
)

var (
	ErrTimeoutLimitExceeded = errors.New("worker shutdown take too much time")
)

type TaskProducer interface {
	ProduceTask(ctx context.Context, tasks []*models.TaskPredictionData) error
	ProduceTaskDelete(ctx context.Context, tasksIds []int64) error
}

type TasksRepository interface {
	GetTasksToProduce(ctx context.Context) ([]*models.TaskPredictionData, error)
	GetTasksToDelete(ctx context.Context) ([]models.TaskPredictionData, error)
	MarkOutbox(ctx context.Context, outboxIDs []int64) error
}

type runDetails struct {
	closeFunc  func()
	closeSyncM *sync.Mutex
}

type Worker struct {
	producer   TaskProducer
	repository TasksRepository
	rd         *runDetails
	log        *slog.Logger
}

func New(log *slog.Logger, producer TaskProducer, repository TasksRepository) *Worker {
	details := &runDetails{
		closeSyncM: &sync.Mutex{},
	}
	return &Worker{
		log:        log,
		producer:   producer,
		repository: repository,
		rd:         details,
	}
}

func (w *Worker) MustRun() {
	ctx, cancel := context.WithCancel(context.Background())
	w.rd.closeFunc = cancel
	w.Run(ctx)
}

func (w *Worker) Close() error {
	w.rd.closeSyncM.Lock()
	w.rd.closeFunc()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	okCh := make(chan struct{})

	// unlock happy path
	go func() {
		w.rd.closeSyncM.Lock()
		okCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		return ErrTimeoutLimitExceeded
	case <-okCh:
	}

	return nil
}

func (w *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Millisecond * 100)

	for {
		select {
		case <-ctx.Done():
			w.rd.closeSyncM.Unlock()
			return
		case <-ticker.C:
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				if err := w.ExtractAndProduceCreate(ctx); err != nil {
					w.log.Error("error while producing created tasks", sl.Err(err))
				}
				wg.Done()
			}()
			go func() {
				if err := w.ExtractAndProduceDelete(ctx); err != nil {
					w.log.Error("error while producing deleted tasks", sl.Err(err))
				}
				wg.Done()
			}()
			wg.Wait()
		}
	}
}

func (w *Worker) ExtractAndProduceCreate(ctx context.Context) error {
	tasks, err := w.repository.GetTasksToProduce(ctx)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return nil
	}

	if err := w.producer.ProduceTask(ctx, tasks); err != nil {
		return err
	}

	ids := make([]int64, len(tasks))
	for i, taskPD := range tasks {
		ids[i] = taskPD.OutboxID
	}

	if err := w.repository.MarkOutbox(ctx, ids); err != nil {
		return err
	}

	return nil
}

func (w *Worker) ExtractAndProduceDelete(ctx context.Context) error {
	tasksDetails, err := w.repository.GetTasksToDelete(ctx)
	if err != nil {
		return err
	}

	details := make([]int64, len(tasksDetails))
	for i, taskDetails := range tasksDetails {
		details[i] = taskDetails.ID
	}

	if err = w.producer.ProduceTaskDelete(ctx, details); err != nil {
		return err
	}

	for i, taskDetails := range tasksDetails {
		details[i] = taskDetails.OutboxID
	}

	if err := w.repository.MarkOutbox(ctx, details); err != nil {
		return err
	}

	return nil
}
