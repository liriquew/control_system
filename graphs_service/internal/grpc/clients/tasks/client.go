package tasksclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	"github.com/liriquew/graphs_service/internal/lib/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type TasksClient struct {
	client tsks_pb.TasksClient
	log    *slog.Logger
}

func NewTasksClient(log *slog.Logger, cfg config.ClientConfig) (*TasksClient, error) {
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

	return &TasksClient{
		client: tsks_pb.NewTasksClient(cc),
		log:    log,
	}, nil
}

func InterceptorLogger(log *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		log.Log(ctx, slog.Level(level), msg, fields...)
	})
}

var (
	ErrInternal = errors.New("internal")
	ErrNotFound = errors.New("not found")
)

func (tc *TasksClient) GetPredictedTasks(ctx context.Context, taskIDs []int64) ([]*tsks_pb.PredictedTask, []int64, error) {
	resp, err := tc.client.GetPredictedTasks(ctx, &tsks_pb.TasksIDs{
		IDs: taskIDs,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, nil, ErrNotFound
			}
		}

		return nil, nil, fmt.Errorf("%w: %w", ErrInternal, err)
	}

	return resp.Tasks, resp.UnpredictedUIDs, nil
}

func (tc *TasksClient) TaskExists(ctx context.Context, taskID, groupID int64) error {
	_, err := tc.client.TaskExists(ctx, &tsks_pb.TaskExistsRequest{
		TaskID:  taskID,
		GroupID: groupID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return ErrNotFound
			}
		}

		return fmt.Errorf("%w: %w", ErrInternal, err)
	}

	return nil
}
