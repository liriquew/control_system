package app

import (
	"fmt"
	"log/slog"

	authclient "github.com/liriquew/tasks_service/internal/grpc/clients/auth_client"
	grphclient "github.com/liriquew/tasks_service/internal/grpc/clients/graphs_client"
	grpsclient "github.com/liriquew/tasks_service/internal/grpc/clients/groups_client"
	predictionsclient "github.com/liriquew/tasks_service/internal/grpc/clients/predictions_client"
	"github.com/liriquew/tasks_service/internal/kafka"
	"github.com/liriquew/tasks_service/internal/lib/config"
	"github.com/liriquew/tasks_service/internal/service/tasks"
	"github.com/liriquew/tasks_service/pkg/logger/sl"

	grpcapp "github.com/liriquew/tasks_service/internal/app/grpc_app"
	tasks_repository "github.com/liriquew/tasks_service/internal/repository"
)

type App struct {
	GRPCServer *grpcapp.App
	closers    []func() error
	log        *slog.Logger
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

	authClient, err := authclient.NewAuthClient(log, cfg.AuthClient)
	if err != nil {
		panic(err)
	}

	prdtClient, err := predictionsclient.NewPredictionsClient(log, cfg.PredictionsClient)
	if err != nil {
		panic(err)
	}

	grpClient, err := grpsclient.NewGroupsClient(log, cfg.GroupsClient)
	if err != nil {
		panic(err)
	}

	grphClient, err := grphclient.NewGraphClient(log, cfg.GraphsClient)
	if err != nil {
		panic(err)
	}

	tasksService := tasks.NewServerAPI(log, storage, producer, authClient, prdtClient, grpClient, grphClient)

	app := grpcapp.New(log, tasksService, cfg.TasksService.Port)

	mainApp := &App{GRPCServer: app, log: log}
	mainApp.closers = append(mainApp.closers, storage.Close)
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
