package outbox

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/liriquew/control_system/tasks_service/internal/models"
	"github.com/liriquew/control_system/tasks_service/pkg/logger/sl"
)

var (
	ErrTimeoutLimitExceeded = errors.New("worker shutdown take too much time")
)

// интерфейс для отправки сообщений
type TaskProducer interface {
	ProduceTask(ctx context.Context, tasks []*models.TaskPredictionData) error
	ProduceTaskDelete(ctx context.Context, tasksIds []int64) error
}

// интерфейс для извлечения обновленных записей
type TasksRepository interface {
	GetTasksToProduce(ctx context.Context) ([]*models.TaskPredictionData, error)
	GetTasksToDelete(ctx context.Context) ([]models.TaskPredictionData, error)
	MarkOutbox(ctx context.Context, outboxIDs []int64) error
}

type runDetails struct {
	// функция для остановки процесса
	// извлечения и публикации сообщений
	closeFunc func()
	// для синхронизации
	// блокируется при завершении (Worker.Close())
	// разблокируется в функции Worker.Run() при обнаружении отмены контекста
	closeSyncM *sync.Mutex
}

// структура описывающая механизм извелечения измененных задач
// и публикующая сообщения в очередь
type Worker struct {
	producer   TaskProducer
	repository TasksRepository
	rd         *runDetails
	log        *slog.Logger
}

// New - конструктор, возвращает экземпляр Worker
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

// запускает процесс извлечения изменений задач из БД,
// с последующей публикацией в очередь
func (w *Worker) MustRun() {
	ctx, cancel := context.WithCancel(context.Background())
	// сохраняем функцию для отмены контекста, для корректной остановки
	w.rd.closeFunc = cancel
	// используем созданный контекст для запуска
	w.Run(ctx)
}

func (w *Worker) Close() error {
	w.rd.closeSyncM.Lock()
	// отменяем контекст
	w.rd.closeFunc()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	okCh := make(chan struct{})

	// разблокировка, в случае,
	// если процесс извлечения и публикации остановился
	go func() {
		w.rd.closeSyncM.Lock()
		okCh <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		// превышено время ожидания
		return ErrTimeoutLimitExceeded
	case <-okCh:
		// случай корректной остановки
	}

	return nil
}

// Run - запускает процесс извлечения измененных задач с последующей публикацией
func (w *Worker) Run(ctx context.Context) {
	// каждые 100 мс проверяем БД
	ticker := time.NewTicker(time.Millisecond * 100)

	for {
		select {
		case <-ctx.Done():
			// если контекст отменен, значит необходимо остановиться,
			// разблокировка мьютекса будет сигнализировать об успешной остановке
			w.rd.closeSyncM.Unlock()
			return
		case <-ticker.C:
			wg := sync.WaitGroup{}
			wg.Add(2)
			// параллельно извлекаем информацию для разных топиков
			go func() {
				// задачи необходимо передать в сервис прогнозирования
				if err := w.ExtractAndProduceCreate(ctx); err != nil {
					w.log.Error("error while producing created tasks", sl.Err(err))
				}
				wg.Done()
			}()
			go func() {
				// задача удалена, необходимо уведомить сервис прогнозирования
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
	// извлекаем завершенные задачи
	tasks, err := w.repository.GetTasksToProduce(ctx)
	if err != nil {
		return err
	}

	if len(tasks) == 0 {
		return nil
	}

	// публикуем информацию о задачах
	if err := w.producer.ProduceTask(ctx, tasks); err != nil {
		return err
	}

	ids := make([]int64, len(tasks))
	for i, taskPD := range tasks {
		ids[i] = taskPD.OutboxID
	}

	// сохраняем информацию о том, что задачи были опубликованы
	if err := w.repository.MarkOutbox(ctx, ids); err != nil {
		return err
	}

	return nil
}

func (w *Worker) ExtractAndProduceDelete(ctx context.Context) error {
	// извлекаем информацию о том, какие задачи были удалены
	tasksDetails, err := w.repository.GetTasksToDelete(ctx)
	if err != nil {
		return err
	}

	details := make([]int64, len(tasksDetails))
	for i, taskDetails := range tasksDetails {
		details[i] = taskDetails.ID
	}

	// публикуем информацию о том, что какие-то задачи были удалены
	if err = w.producer.ProduceTaskDelete(ctx, details); err != nil {
		return err
	}

	for i, taskDetails := range tasksDetails {
		details[i] = taskDetails.OutboxID
	}

	// отмечаем, что информация об удалении задач была опубликована
	if err := w.repository.MarkOutbox(ctx, details); err != nil {
		return err
	}

	return nil
}
