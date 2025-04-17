package app

import (
	"fmt"
	"log/slog"

	"github.com/liriquew/graphs_service/internal/lib/config"
	graphsservice "github.com/liriquew/graphs_service/internal/service/graphs"
	"github.com/liriquew/graphs_service/pkg/logger/sl"

	authclient "github.com/liriquew/graphs_service/internal/grpc/clients/auth"
	tasksclient "github.com/liriquew/graphs_service/internal/grpc/clients/tasks"

	grpcapp "github.com/liriquew/graphs_service/internal/app/grpc_app"
	auth_repository "github.com/liriquew/graphs_service/internal/repository"
)

type App struct {
	GRPCServer *grpcapp.App
	closers    []func() error
	log        *slog.Logger
}

func New(log *slog.Logger, cfg config.AppConfig) *App {
	storage, err := auth_repository.NewGraphsRepository(cfg.Storage)
	if err != nil {
		panic(err)
	}

	authClient, err := authclient.NewAuthClient(log, cfg.AuthClient)
	if err != nil {
		panic(err)
	}

	tasksClient, err := tasksclient.NewTasksClient(log, cfg.TasksClient)
	if err != nil {
		panic(err)
	}

	graphsService := graphsservice.NewServerAPI(log, storage, authClient, tasksClient)

	app := grpcapp.New(log, graphsService, cfg.ServiceConfig.Port)

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
