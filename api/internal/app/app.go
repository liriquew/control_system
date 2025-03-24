package app

import (
	"log"

	"github.com/liriquew/control_system/internal/api_handlers/auth"
	"github.com/liriquew/control_system/internal/api_handlers/graphs"
	group_handlers "github.com/liriquew/control_system/internal/api_handlers/groups"
	"github.com/liriquew/control_system/internal/api_handlers/task"
	appapi "github.com/liriquew/control_system/internal/app/api"
	"github.com/liriquew/control_system/internal/config"
	predictions_client "github.com/liriquew/control_system/internal/grpc/client"

	graphstasks_repository "github.com/liriquew/control_system/internal/repository/graph_tasks"
	graphs_repository "github.com/liriquew/control_system/internal/repository/graphs"
	group_repository "github.com/liriquew/control_system/internal/repository/groups"
	tasks_repository "github.com/liriquew/control_system/internal/repository/tasks"
	users_repository "github.com/liriquew/control_system/internal/repository/users"
	"github.com/liriquew/control_system/internal/storage"

	graphs_service "github.com/liriquew/control_system/internal/service/graphs"
	groups_service "github.com/liriquew/control_system/internal/service/groups"
	tasks_service "github.com/liriquew/control_system/internal/service/tasks"
	users_service "github.com/liriquew/control_system/internal/service/users"

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
	graphsTasksRepository, err := graphstasks_repository.NewGraphsTasksRepository(db)
	if err != nil {
		panic(err)
	}

	predictorClient, err := predictions_client.New(infoLog, &cfg.PredictionsService)
	if err != nil {
		panic(err)
	}

	userService, err := users_service.NewUserService(userRepository, infoLog, errorLog)
	if err != nil {
		panic(err)
	}
	taskService, err := tasks_service.NewTaskService(taskRepository, predictorClient, infoLog, errorLog)
	if err != nil {
		panic(err)
	}
	groupService, err := groups_service.NewGroupService(groupRepository, infoLog, errorLog)
	if err != nil {
		panic(err)
	}
	graphService, err := graphs_service.NewGraphsService(graphReository, predictorClient, graphsTasksRepository, infoLog, errorLog)
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
