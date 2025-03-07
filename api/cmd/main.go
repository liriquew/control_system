package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"time_manage/internal/app"
	"time_manage/internal/config"
	"time_manage/internal/storage"
)

type API struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	storage  *storage.Storage
}

func main() {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ltime)

	cfg := config.MustLoad()

	r := app.New(infoLog, errorLog, cfg)

	srv := &http.Server{
		Addr:    ":" + "8080",
		Handler: r.Router,
	}

	infoLog.Println("RUN SERVER: localhost:8080")

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errorLog.Println("listen", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	infoLog.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		errorLog.Println("Server forced to shutdown:", err)
	}

	infoLog.Println("Server exiting")
}
