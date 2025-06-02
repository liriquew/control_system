package app

import (
	"log/slog"

	"github.com/liriquew/control_system/api/internal/api_handlers/auth"
	"github.com/liriquew/control_system/api/internal/api_handlers/graphs"
	"github.com/liriquew/control_system/api/internal/api_handlers/groups"
	tasks "github.com/liriquew/control_system/api/internal/api_handlers/task"
	appapi "github.com/liriquew/control_system/api/internal/app/api"
	authclient "github.com/liriquew/control_system/api/internal/grpc/clients/auth"
	graphsclient "github.com/liriquew/control_system/api/internal/grpc/clients/graphs"
	groupsclient "github.com/liriquew/control_system/api/internal/grpc/clients/groups"
	predictionsclient "github.com/liriquew/control_system/api/internal/grpc/clients/predictions"
	tasksclient "github.com/liriquew/control_system/api/internal/grpc/clients/tasks"
	"github.com/liriquew/control_system/api/internal/lib/config"

	"github.com/go-chi/chi/v5"
)

type App struct {
	Router *chi.Mux
}

type Dependencies struct {
}

func New(log *slog.Logger, cfg config.AppConfig) *App {
	authClient, err := authclient.NewAuthClient(log, cfg.AuthClient)
	if err != nil {
		panic(err)
	}
	tasksClient, err := tasksclient.NewTasksClient(log, cfg.TasksService)
	if err != nil {
		panic(err)
	}
	groupsClient, err := groupsclient.NewGroupsClient(log, cfg.GroupsClient)
	if err != nil {
		panic(err)
	}
	graphsClient, err := graphsclient.NewGraphClient(log, cfg.GraphsClient)
	if err != nil {
		panic(err)
	}
	predictionsClient, err := predictionsclient.NewPredictionsClient(log, cfg.PredictionsClient)
	if err != nil {
		panic(err)
	}

	auth := auth.New(log, authClient)

	tasks := tasks.New(log, tasksClient, groupsClient, predictionsClient)

	groups := groups.New(log, groupsClient, tasksClient, authClient)

	graphs := graphs.New(log, graphsClient, tasksClient)

	r := appapi.New(auth, tasks, groups, graphs)

	return &App{r}
}
