package app

import (
	"fmt"
	"log/slog"

	"github.com/liriquew/control_system/graphs_service/internal/lib/config"
	graphs_service "github.com/liriquew/control_system/graphs_service/internal/service/graphs"
	"github.com/liriquew/control_system/graphs_service/pkg/logger/sl"

	tasks_client "github.com/liriquew/control_system/graphs_service/internal/grpc/clients/tasks"

	grpcapp "github.com/liriquew/control_system/graphs_service/internal/app/grpc_app"
	"github.com/liriquew/control_system/graphs_service/internal/repository"
)

type App struct {
	GRPCServer *grpcapp.App
	closers    []func() error
	log        *slog.Logger
}

func New(log *slog.Logger, cfg config.AppConfig) *App {
	storage, err := repository.New(cfg.Storage)
	if err != nil {
		panic(err)
	}

	tasksClient, err := tasks_client.New(log, cfg.TasksClient)
	if err != nil {
		panic(err)
	}

	service := graphs_service.New(log, storage, tasksClient)

	app := grpcapp.New(log, service, cfg.ServiceConfig.Port)

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
