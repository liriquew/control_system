package app

import (
	"log"
	"time_manage/internal/api_handlers/auth"
	"time_manage/internal/api_handlers/graphs"
	group_handlers "time_manage/internal/api_handlers/groups"
	"time_manage/internal/api_handlers/task"
	appapi "time_manage/internal/app/api"
	"time_manage/internal/config"
	predictions_client "time_manage/internal/grpc/client"

	graphs_repository "time_manage/internal/repository/graphs"
	group_repository "time_manage/internal/repository/groups"
	tasks_repository "time_manage/internal/repository/tasks"
	users_repository "time_manage/internal/repository/users"
	"time_manage/internal/storage"

	graphs_service "time_manage/internal/service/graphs"
	groups_service "time_manage/internal/service/groups"
	tasks_service "time_manage/internal/service/tasks"
	users_service "time_manage/internal/service/users"

	"github.com/go-chi/chi/v5"
)

type App struct {
	Router *chi.Mux
}

func New(infoLog, errorLog *log.Logger, cfg config.AppConfig) *App {
	db, err := storage.NewStorage(&cfg.Storage)
	if err != nil {
		panic(err)
	}

	userRepository, err := users_repository.NewUserRepository(db)
	if err != nil {
		panic(err)
	}
	taskRepository, err := tasks_repository.NewTaskRepository(db)
	if err != nil {
		panic(err)
	}
	groupRepository, err := group_repository.NewGroupsRepository(db)
	if err != nil {
		panic(err)
	}
	graphReository, err := graphs_repository.NewGraphsRepository(db)
	if err != nil {
		panic(err)
	}

	taskClient, err := predictions_client.New(infoLog, &cfg.PredictionsService)
	if err != nil {
		panic(err)
	}

	userService, err := users_service.NewUserService(userRepository, infoLog, errorLog)
	if err != nil {
		panic(err)
	}
	taskService, err := tasks_service.NewTaskService(taskRepository, taskClient, infoLog, errorLog)
	if err != nil {
		panic(err)
	}
	groupService, err := groups_service.NewGroupService(groupRepository, infoLog, errorLog)
	if err != nil {
		panic(err)
	}
	graphService, err := graphs_service.NewGraphsService(graphReository, infoLog, errorLog)
	if err != nil {
		panic(err)
	}

	auth := auth.New(infoLog, errorLog, userService)

	tasks := task.New(infoLog, errorLog, taskService)

	groups := group_handlers.New(groupService, infoLog, errorLog)

	graphs := graphs.New(graphService, infoLog, errorLog)

	r := appapi.New(auth, tasks, groups, graphs)

	return &App{r}
}
