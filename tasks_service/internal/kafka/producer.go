package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/liriquew/control_system/tasks_service/internal/lib/config"
	"github.com/liriquew/control_system/tasks_service/internal/models"
	"github.com/liriquew/control_system/tasks_service/pkg/logger/sl"
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
		BatchSize:     1000,
		BatchTimeout:  100 * time.Millisecond,
		QueueCapacity: 10000,
		MaxAttempts:   2,
	})
	deleteWriter := kafka.NewWriter(kafka.WriterConfig{
		Brokers:       []string{cfg.ConnStr},
		Topic:         cfg.DeleteTopic,
		Balancer:      &kafka.LeastBytes{},
		RequiredAcks:  1,
		BatchSize:     1000,
		BatchTimeout:  100 * time.Millisecond,
		QueueCapacity: 10000,
		MaxAttempts:   2,
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

func (p *ProducerWrapper) ProduceTask(ctx context.Context, tasks []*models.TaskPredictionData) error {
	messages := make([]kafka.Message, len(tasks))
	for i, task := range tasks {
		message, _ := json.Marshal(task)
		p.log.Debug("task to sent", slog.Any("task", task))
		messages[i] = kafka.Message{
			Key:   convertInt64ByteArr(task.ID),
			Value: message,
		}
	}
	err := p.w.WriteMessages(ctx, messages...)
	if err != nil {
		p.log.Error("failed to send task to Kafka", sl.Err(err))
		return fmt.Errorf("write message: %w", err)
	}
	// p.log.Debug("create tasks messages was sent")

	return nil
}

func (p *ProducerWrapper) ProduceTaskDelete(ctx context.Context, tasksIds []int64) error {
	messages := make([]kafka.Message, len(tasksIds))
	for i, taskId := range tasksIds {
		message, _ := json.Marshal(models.Task{ID: taskId})
		messages[i] = kafka.Message{
			Key:   convertInt64ByteArr(taskId),
			Value: message,
		}
	}

	err := p.deleteW.WriteMessages(ctx, messages...)
	if err != nil {
		p.log.Error("failed to send task to Kafka", sl.Err(err))
		return fmt.Errorf("write message: %w", err)
	}
	// p.log.Debug("delete messages was sent")

	return nil
}
