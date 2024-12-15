package app

import (
	"log"
	"time_manage/internal/api_handlers/auth"
	"time_manage/internal/api_handlers/task"
	appapi "time_manage/internal/app/api"
	predictions_client "time_manage/internal/grpc/client"
	"time_manage/internal/storage"

	"github.com/go-chi/chi/v5"
)

type App struct {
	Router *chi.Mux
}

func New(infoLog, errorLog *log.Logger) *App {
	storage, err := storage.New()
	if err != nil {
		panic(err)
	}

	auth := auth.New(infoLog, errorLog, storage)

	taskClient, err := predictions_client.New(infoLog)
	if err != nil {
		panic(err)
	}

	tasks := task.New(infoLog, errorLog, storage, taskClient)

	r := appapi.New(auth, tasks)

	return &App{r}
}
