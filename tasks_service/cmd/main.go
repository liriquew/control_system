package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/liriquew/control_system/tasks_service/internal/app"
	"github.com/liriquew/control_system/tasks_service/internal/lib/config"
	"github.com/liriquew/control_system/tasks_service/pkg/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupPrettySlog("Tasks")

	log.Info("", slog.Any("CONFIG", cfg))

	application := app.New(log, cfg)

	go func() {
		application.GRPCServer.MustRun()
	}()

	go func() {
		application.OutboxMachine.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	application.Stop()
	log.Info("Gracefully stopped")
}
