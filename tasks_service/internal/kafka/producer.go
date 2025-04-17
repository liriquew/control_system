package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/liriquew/tasks_service/internal/lib/config"
	"github.com/liriquew/tasks_service/internal/models"
	"github.com/liriquew/tasks_service/pkg/logger/sl"
	"github.com/segmentio/kafka-go"
)

type ProducerWrapper struct {
	w       *kafka.Writer
	deleteW *kafka.Writer
	log     *slog.Logger
}

func NewProducer(log *slog.Logger, cfg config.KafkaTasksTopicConfig) (*ProducerWrapper, error) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:       []string{cfg.ConnStr},
		Topic:         cfg.Topic,
		Balancer:      &kafka.LeastBytes{},
		RequiredAcks:  1,
		BatchSize:     1000,                   // Увеличить размер батча
		BatchTimeout:  100 * time.Millisecond, // Чаще отправлять батчи
		QueueCapacity: 10000,                  // Увеличить размер очереди
		MaxAttempts:   2,                      // Уменьшить число попыток
	})
	deleteWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:       []string{cfg.ConnStr},
		Topic:         cfg.DeleteTopic,
		Balancer:      &kafka.LeastBytes{},
		RequiredAcks:  1,
		BatchSize:     1000,                   // Увеличить размер батча
		BatchTimeout:  100 * time.Millisecond, // Чаще отправлять батчи
		QueueCapacity: 10000,                  // Увеличить размер очереди
		MaxAttempts:   2,                      // Уменьшить число попыток
	})

	return &ProducerWrapper{
		w:       writer,
		deleteW: deleteWriter,
		log:     log,
	}, nil
}

func (p *ProducerWrapper) Close() error {
	if err := p.w.Close(); err != nil {
		p.log.Error("failed to close Kafka producer", sl.Err(err))
		return err
	}
	return nil
}

func (p *ProducerWrapper) ProduceTaskPredictionData(ctx context.Context, task *models.TaskPredictionData) error {
	message, err := json.Marshal(task)
	if err != nil {
		p.log.Error("failed to marshal task", sl.Err(err))
		return fmt.Errorf("marshal task: %w", err)
	}

	err = p.w.WriteMessages(ctx,
		kafka.Message{
			Key:   convertInt64ByteArr(task.ID),
			Value: message,
		},
	)
	if err != nil {
		p.log.Error("failed to send task to Kafka", sl.Err(err))
		return fmt.Errorf("write message: %w", err)
	}

	p.log.Debug("task sent to Kafka", slog.Int64("task_id", task.ID))
	return nil
}

func (p *ProducerWrapper) ProduceTaskPredictionDataDelete(ctx context.Context, taskID int64) error {
	message, err := json.Marshal(&models.TaskPredictionData{
		ID: taskID,
	})
	if err != nil {
		p.log.Error("failed to marshal task", sl.Err(err))
		return fmt.Errorf("marshal task: %w", err)
	}

	err = p.deleteW.WriteMessages(ctx,
		kafka.Message{
			Key:   convertInt64ByteArr(taskID),
			Value: message,
		},
	)
	if err != nil {
		p.log.Error("failed to send task to Kafka", sl.Err(err))
		return fmt.Errorf("write message: %w", err)
	}

	p.log.Debug("task sent to Kafka", slog.Int64("task_id", taskID))
	return nil
}
