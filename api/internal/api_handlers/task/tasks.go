package task

import (
	"log/slog"

	groupsclient "github.com/liriquew/control_system/internal/grpc/clients/groups"
	predictionsclient "github.com/liriquew/control_system/internal/grpc/clients/predictions"
	tasksclient "github.com/liriquew/control_system/internal/grpc/clients/tasks"
)

type TasksService struct {
	TaskAPI
	TasksMiddleware
}

func New(
	log *slog.Logger,
	client tasksclient.TasksClient,
	groupsClient groupsclient.GroupsClient,
	predictionsClient predictionsclient.PredictionsClient,
) TasksService {
	service := NewTasksService(log, client, groupsClient, predictionsClient)
	middleware := NewTasksMiddleware(log, client)

	return TasksService{
		TaskAPI:         service,
		TasksMiddleware: *middleware,
	}
}
