package app

import (
	"fmt"
	"log/slog"

	predictionsclient "github.com/liriquew/tasks_service/internal/grpc/clients/predictions_client"
	"github.com/liriquew/tasks_service/internal/kafka"
	"github.com/liriquew/tasks_service/internal/lib/config"
	"github.com/liriquew/tasks_service/internal/outbox"
	"github.com/liriquew/tasks_service/internal/service/tasks"
	"github.com/liriquew/tasks_service/pkg/logger/sl"

	grpcapp "github.com/liriquew/tasks_service/internal/app/grpc_app"
	tasks_repository "github.com/liriquew/tasks_service/internal/repository"
)

type App struct {
	GRPCServer    *grpcapp.App
	OutboxMachine *outbox.Worker
	closers       []func() error
	log           *slog.Logger
}

func New(log *slog.Logger, cfg config.AppConfig) *App {
	storage, err := tasks_repository.NewTaskRepository(cfg.Storage)
	if err != nil {
		panic(err)
	}

	producer, err := kafka.NewProducer(log, cfg.KafkaConfig)
	if err != nil {
		panic(err)
	}

	prdtClient, err := predictionsclient.NewPredictionsClient(log, cfg.PredictionsClient)
	if err != nil {
		panic(err)
	}

	outbox := outbox.New(log, producer, storage)

	tasksService := tasks.NewServerAPI(log, storage, prdtClient)

	app := grpcapp.New(log, tasksService, cfg.TasksService.Port)

	mainApp := &App{GRPCServer: app, log: log, OutboxMachine: outbox}
	mainApp.closers = append(mainApp.closers, storage.Close, outbox.Close)
	return mainApp
}

func (a *App) Stop() {
	const op = "app.App.Stop"

	for _, c := range a.closers {
		if err := c(); err != nil {
			a.log.Warn("ERROR", sl.Err(fmt.Errorf("%s: %w", op, err)))
		}
	}

	a.GRPCServer.Stop()
}
