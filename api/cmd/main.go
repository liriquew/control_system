package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/liriquew/control_system/internal/app"
	"github.com/liriquew/control_system/internal/lib/config"
	"github.com/liriquew/control_system/pkg/logger"
	"github.com/liriquew/control_system/pkg/logger/sl"
)

func main() {
	log := logger.SetupPrettySlog("API_GATEWAY")

	cfg := config.MustLoad()

	log.Info("", slog.Any("CONFIG", cfg))

	r := app.New(log, cfg)

	srv := &http.Server{
		Addr:    ":" + "8080",
		Handler: r.Router,
	}

	log.Info("RUN SERVER: localhost:8080")

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("listen", sl.Err(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Info("Server forced to shutdown:", sl.Err(err))
	}

	log.Info("Server exiting")
}
