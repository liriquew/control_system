package groupsclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	gprs_pb "github.com/liriquew/control_system/services_protos/groups_service"
	"github.com/liriquew/graphs_service/internal/lib/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GroupsClient struct {
	client gprs_pb.GroupsClient
	log    *slog.Logger
}

func NewGroupsClient(log *slog.Logger, cfg config.ClientConfig) (*GroupsClient, error) {
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

	return &GroupsClient{
		client: gprs_pb.NewGroupsClient(cc),
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

func (g *GroupsClient) CheckAdminPermission(ctx context.Context, userID, groupID int64) error {
	_, err := g.client.CheckAdminPermission(ctx, &gprs_pb.GroupMember{
		UserID:  userID,
		GroupID: groupID,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return ErrDeny
		}

		return fmt.Errorf("internal error: %w", err)
	}

	return nil
}

func (g *GroupsClient) CheckEditorPermission(ctx context.Context, userID, groupID int64) error {
	_, err := g.client.CheckEditorPermission(ctx, &gprs_pb.GroupMember{
		UserID:  userID,
		GroupID: groupID,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return ErrDeny
		}

		return fmt.Errorf("internal error: %w", err)
	}

	return nil
}

func (g *GroupsClient) CheckMemberPermission(ctx context.Context, userID, groupID int64) error {
	_, err := g.client.CheckMemberPermission(ctx, &gprs_pb.GroupMember{
		UserID:  userID,
		GroupID: groupID,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
			return ErrDeny
		}

		return fmt.Errorf("internal error: %w", err)

	}

	return nil
}
