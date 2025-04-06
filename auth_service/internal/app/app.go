package app

import (
	"fmt"
	"log/slog"

	"github.com/liriquew/auth_service/internal/lib/config"
	"github.com/liriquew/auth_service/internal/service/auth"
	"github.com/liriquew/auth_service/pkg/logger/sl"

	grpcapp "github.com/liriquew/auth_service/internal/app/grpc_app"
	auth_repository "github.com/liriquew/auth_service/internal/repository"
)

type App struct {
	GRPCServer *grpcapp.App
	closers    []func() error
	log        *slog.Logger
}

func New(log *slog.Logger, cfg config.AppConfig) *App {
	storage, err := auth_repository.NewAuthRepository(cfg.Storage)
	if err != nil {
		panic(err)
	}

	tasksService := auth.NewServerAPI(log, storage)

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
