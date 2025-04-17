package authclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	"github.com/liriquew/control_system/internal/lib/config"
	"github.com/liriquew/control_system/internal/models"
	"github.com/liriquew/control_system/pkg/logger/sl"
	auth_pb "github.com/liriquew/control_system/services_protos/auth_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPCAuthClient struct {
	client auth_pb.AuthClient
	log    *slog.Logger
}

func NewAuthClient(log *slog.Logger, cfg config.ClientConfig) (*GRPCAuthClient, error) {
	const op = "authclient.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(cfg.Retries)),
		grpcretry.WithPerRetryTimeout(cfg.Timeout),
	}

	logOpts := []grpclog.Option{
		grpclog.WithLogOnEvents(grpclog.PayloadReceived, grpclog.PayloadSent),
	}

	address := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	cc, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &GRPCAuthClient{
		client: auth_pb.NewAuthClient(cc),
		log:    log,
	}, nil
}

func InterceptorLogger(log *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		log.Log(ctx, slog.Level(level), msg, fields...)
	})
}

var (
	ErrMissedJWT          = errors.New("missed jwt token")
	ErrDeny               = errors.New("access denied")
	ErrInvalidCredentials = errors.New("err bad params")
	ErrAlreadyExists      = errors.New("err already exists")
	ErrUserNotFound       = errors.New("not found")
	ErrUnauthenticated    = errors.New("Unauthenticated")
)

type AuthClient interface {
	SignIn(ctx context.Context, creds *models.User) (string, error)
	SignUp(ctx context.Context, creds *models.User) (string, error)
	Authenticate(ctx context.Context, jwtToken string) (int64, error)
	GetUsersDetails(ctx context.Context, userIDs []int64) ([]*models.User, error)
}

func (a *GRPCAuthClient) SignIn(ctx context.Context, creds *models.User) (string, error) {
	a.log.Debug("send request to sign in")
	resp, err := a.client.SignIn(ctx, &auth_pb.UserCredentials{
		Username: creds.Username,
		Password: creds.Password,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return "", ErrInvalidCredentials
			case codes.NotFound:
				return "", ErrUserNotFound
			case codes.Unauthenticated:
				return "", ErrInvalidCredentials
			}
		}

		a.log.Error("failed to sign in", sl.Err(err))
		return "", err
	}

	return resp.JWT, nil
}

func (a *GRPCAuthClient) SignUp(ctx context.Context, creds *models.User) (string, error) {
	a.log.Debug("send request to sign up")
	resp, err := a.client.SignUp(ctx, &auth_pb.UserCredentials{
		Username: creds.Username,
		Password: creds.Password,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return "", ErrInvalidCredentials
			case codes.AlreadyExists:
				return "", ErrAlreadyExists
			}
		}

		a.log.Error("failed to sign up", sl.Err(err))
		return "", err
	}

	return resp.JWT, nil
}

func (a *GRPCAuthClient) Authenticate(ctx context.Context, jwtToken string) (int64, error) {
	if jwtToken == "" {
		return 0, ErrMissedJWT
	}
	a.log.Debug("send request to auth")
	userID, err := a.client.Authenticate(ctx, &auth_pb.JWT{
		JWT: jwtToken,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.Unauthenticated {
			return 0, ErrDeny
		}

		a.log.Error("failed to authenticate", sl.Err(err))
		return 0, err
	}

	return userID.ID, nil
}

// func (s *serverAPI) GetUsersDetails(ctx context.Context, userIDs *auth_pb.UserIDs) (*auth_pb.ListUserDetails, error) {

func (a *GRPCAuthClient) GetUsersDetails(ctx context.Context, userIDs []int64) ([]*models.User, error) {
	resp, err := a.client.GetUsersDetails(ctx, &auth_pb.UserIDs{
		UserIDs: userIDs,
	})
	if err != nil {
		return nil, err
	}

	return models.ConvertUsersFromProto(resp.Users), nil
}
