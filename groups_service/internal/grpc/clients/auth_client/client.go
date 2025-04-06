package authclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	auth_pb "github.com/liriquew/control_system/services_protos/auth_service"
	"github.com/liriquew/groups_service/internal/lib/config"
	"github.com/liriquew/groups_service/pkg/logger/sl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type AuthClient struct {
	client auth_pb.AuthClient
	log    *slog.Logger
}

func NewAuthClient(log *slog.Logger, cfg config.ClientConfig) (*AuthClient, error) {
	const op = "authclient.New"

	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
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

	return &AuthClient{
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
	ErrMissedJWT = errors.New("missed jwt token")
	ErrDeny      = errors.New("access denied")
)

func (a *AuthClient) Authenticate(ctx context.Context, jwtToken string) (int64, error) {
	if jwtToken == "" {
		return 0, ErrMissedJWT
	}

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
