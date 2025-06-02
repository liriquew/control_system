package auth

import (
	"log/slog"

	authclient "github.com/liriquew/control_system/api/internal/grpc/clients/auth"
)

type AuthService struct {
	AuthAPI
	AuthMiddleware
}

func New(log *slog.Logger, client authclient.AuthClient) AuthService {
	service := NewAuthService(log, client)
	middleware := NewAuthMiddleware(log, client)

	return AuthService{
		AuthAPI:        service,
		AuthMiddleware: *middleware,
	}
}
