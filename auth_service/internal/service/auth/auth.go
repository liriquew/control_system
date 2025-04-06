package auth

import (
	"context"
	"errors"
	"log/slog"

	"github.com/liriquew/auth_service/internal/lib/jwt"
	"github.com/liriquew/auth_service/internal/models"
	"github.com/liriquew/auth_service/internal/repository"
	"github.com/liriquew/auth_service/pkg/logger/sl"
	auth_pb "github.com/liriquew/control_system/services_protos/auth_service"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type tasksRepository interface {
	SaveUser(ctx context.Context, username string, password []byte) (int64, error)
	GetUser(ctx context.Context, username string) (*models.User, error)
}

type serverAPI struct {
	auth_pb.UnimplementedAuthServer
	repository tasksRepository
	log        *slog.Logger
}

func Register(gRPC *grpc.Server, taskServiceAPI auth_pb.AuthServer) {
	auth_pb.RegisterAuthServer(gRPC, taskServiceAPI)
}

func NewServerAPI(log *slog.Logger, taskRepository tasksRepository) *serverAPI {
	return &serverAPI{
		log:        log,
		repository: taskRepository,
	}
}

func (s *serverAPI) SignUp(ctx context.Context, UserCredentials *auth_pb.UserCredentials) (*auth_pb.JWT, error) {
	if UserCredentials.Password == "" || UserCredentials.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "bad params")
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(UserCredentials.Password), bcrypt.MinCost)
	if err != nil {
		s.log.Warn("failed to generate hash", sl.Err(err))
		return nil, status.Error(codes.Internal, "failed to generate hash")
	}

	userID, err := s.repository.SaveUser(ctx, UserCredentials.Username, passHash)
	if err != nil {
		s.log.Error(err.Error())
		if errors.Is(err, repository.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, err
	}

	token, err := jwt.NewToken(userID)
	if err != nil {
		s.log.Error("failed to generate jwt", sl.Err(err))
		return nil, status.Error(codes.Internal, "failed to generate jwt")
	}

	return &auth_pb.JWT{
		JWT: token,
	}, nil
}

func (s *serverAPI) SignIn(ctx context.Context, UserCredentials *auth_pb.UserCredentials) (*auth_pb.JWT, error) {
	if UserCredentials.Password == "" || UserCredentials.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "bad params")
	}

	user, err := s.repository.GetUser(ctx, UserCredentials.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "invalid credentials")
		}
		s.log.Error(err.Error())
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(UserCredentials.Password)); err != nil {
		s.log.Info("invalid credentials", sl.Err(err))
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token, err := jwt.NewToken(user.UID)
	if err != nil {
		s.log.Error("failed to generate jwt", sl.Err(err))
		return nil, status.Error(codes.Internal, "failed to generate jwt")
	}

	return &auth_pb.JWT{
		JWT: token,
	}, nil
}

func (s *serverAPI) Authenticate(ctx context.Context, JWTToken *auth_pb.JWT) (*auth_pb.UserID, error) {
	userID, err := jwt.Validate(JWTToken.JWT)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "failed to validate jwt")
	}

	return &auth_pb.UserID{
		ID: userID,
	}, nil
}
