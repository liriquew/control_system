package groups

import (
	"log/slog"

	authclient "github.com/liriquew/control_system/internal/grpc/clients/auth"
	groupsclient "github.com/liriquew/control_system/internal/grpc/clients/groups"
	tasksclient "github.com/liriquew/control_system/internal/grpc/clients/tasks"
)

type GroupsService struct {
	GroupsAPI
	GroupsMiddleware
}

func New(log *slog.Logger, client groupsclient.GroupsClient, tasksClient tasksclient.TasksClient, authClient authclient.AuthClient) GroupsService {
	service := NewGroupsService(log, client, tasksClient, authClient)
	middleware := NewAuthMiddleware(log, client)

	return GroupsService{
		GroupsAPI:        service,
		GroupsMiddleware: *middleware,
	}
}
