package app

import (
	"fmt"
	"log/slog"

	"github.com/liriquew/groups_service/internal/lib/config"
	"github.com/liriquew/groups_service/internal/service/groups"
	"github.com/liriquew/groups_service/pkg/logger/sl"

	grpcapp "github.com/liriquew/groups_service/internal/app/grpc_app"
	tasks_repository "github.com/liriquew/groups_service/internal/repository"
)

type App struct {
	GRPCServer *grpcapp.App
	closers    []func() error
	log        *slog.Logger
}

func New(log *slog.Logger, cfg config.AppConfig) *App {
	storage, err := tasks_repository.NewGroupRepository(cfg.Storage)
	if err != nil {
		panic(err)
	}

	tasksService := groups.NewServerAPI(log, storage)

	app := grpcapp.New(log, tasksService, cfg.ServiceConfig.Port)

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
