package graphs

import (
	"log/slog"

	graphsclient "github.com/liriquew/control_system/internal/grpc/clients/graphs"
	tasksclient "github.com/liriquew/control_system/internal/grpc/clients/tasks"
)

type GraphsService struct {
	GpraphsAPI
	GraphsMiddleware
}

func New(log *slog.Logger, client graphsclient.GraphClient, tasksClient tasksclient.TasksClient) GraphsService {
	service := NewGraphsService(log, client, tasksClient)
	middleware := NewGraphsMiddleware(log)

	return GraphsService{
		GpraphsAPI:       service,
		GraphsMiddleware: *middleware,
	}
}
