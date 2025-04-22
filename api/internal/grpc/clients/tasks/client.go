package tasksclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	"github.com/liriquew/control_system/internal/entities"
	"github.com/liriquew/control_system/internal/lib/config"
	"github.com/liriquew/control_system/internal/lib/converter"
	"github.com/liriquew/control_system/internal/models"
	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPCTasksClient struct {
	client tsks_pb.TasksClient
	log    *slog.Logger
}

func NewTasksClient(log *slog.Logger, cfg config.ClientConfig) (*GRPCTasksClient, error) {
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

	return &GRPCTasksClient{
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
	ErrInternal         = errors.New("internal")
	ErrNotFound         = errors.New("not found")
	ErrBadParams        = errors.New("")
	ErrPermissionDenied = errors.New("permission denied")
)

type TasksClient interface {
	CreateTask(context.Context, *models.Task) (int64, error)
	GetTask(context.Context, int64) (*models.Task, error)
	GetTaskList(context.Context, int64) ([]*models.Task, error)
	UpdateTask(context.Context, *models.Task) error
	DeleteTask(context.Context, int64) error
	PredictTask(context.Context, int64) (*entities.PredictedTask, error)
	TaskExists(context.Context, int64, int64) error
	GetGroupTasks(context.Context, int64) ([]*models.Task, error)
	GetPredictedTasks(context.Context, []int64) ([]*tsks_pb.PredictedTask, []int64, error)
}

func (c *GRPCTasksClient) CreateTask(ctx context.Context, task *models.Task) (int64, error) {
	resp, err := c.client.CreateTask(ctx, converter.ConvertTaskToProto(task))
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return 0, fmt.Errorf("%w%s", ErrBadParams, st.Message())
			}
		}

		return 0, err
	}

	return resp.ID, nil
}

func (c *GRPCTasksClient) GetTask(ctx context.Context, taskID int64) (*models.Task, error) {
	resp, err := c.client.GetTask(ctx, &tsks_pb.TaskID{
		ID: taskID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.PermissionDenied:
				return nil, ErrPermissionDenied
			case codes.NotFound:
				return nil, ErrNotFound
			}
		}

		return nil, err
	}

	return converter.ConvertTaskToModel(resp), nil
}

func (c *GRPCTasksClient) GetTaskList(ctx context.Context, offset int64) ([]*models.Task, error) {
	resp, err := c.client.GetTaskList(ctx, &tsks_pb.TaskListRequest{
		Offset: offset,
	})
	if err != nil {
		return nil, err
	}

	return converter.ConvertTasksToModel(resp.Tasks), nil
}

func (c *GRPCTasksClient) UpdateTask(ctx context.Context, task *models.Task) error {
	_, err := c.client.UpdateTask(ctx, converter.ConvertTaskToProto(task))
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.PermissionDenied:
				return ErrPermissionDenied
			case codes.FailedPrecondition:
				return fmt.Errorf("%w%s", ErrBadParams, st.Message())
			case codes.InvalidArgument:
				return fmt.Errorf("%w%s", ErrBadParams, st.Message())
			case codes.NotFound:
				return ErrNotFound
			}
		}

		return err
	}

	return nil
}

func (c *GRPCTasksClient) DeleteTask(ctx context.Context, taskID int64) error {
	_, err := c.client.DeleteTask(ctx, &tsks_pb.TaskID{
		ID: taskID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.PermissionDenied:
				return ErrPermissionDenied
			case codes.FailedPrecondition:
				return fmt.Errorf("%w%s", ErrBadParams, st.Message())
			case codes.NotFound:
				return ErrNotFound
			}
		}

		return err
	}

	return nil
}

func (c *GRPCTasksClient) PredictTask(ctx context.Context, taskID int64) (*entities.PredictedTask, error) {
	resp, err := c.client.PredictTask(ctx, &tsks_pb.TaskID{
		ID: taskID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, ErrNotFound
			case codes.PermissionDenied:
				return nil, ErrPermissionDenied
			case codes.FailedPrecondition:
				return nil, fmt.Errorf("%w%s", ErrBadParams, st.Message())
			}
		}

		return nil, err
	}

	return converter.ConvertPredictedTaskToModel(resp), nil
}

func (c *GRPCTasksClient) TaskExists(ctx context.Context, taskID, groupID int64) error {
	_, err := c.client.TaskExists(ctx, &tsks_pb.TaskExistsRequest{
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

func (c *GRPCTasksClient) GetGroupTasks(ctx context.Context, groupID int64) ([]*models.Task, error) {
	resp, err := c.client.GetGroupTasks(ctx, &tsks_pb.GroupID{
		ID: groupID,
	})
	if err != nil {
		return nil, err
	}
	return converter.ConvertTasksToModel(resp.Tasks), nil
}

func (c *GRPCTasksClient) GetPredictedTasks(ctx context.Context, taskIDs []int64) ([]*tsks_pb.PredictedTask, []int64, error) {
	resp, err := c.client.GetPredictedTasks(ctx, &tsks_pb.TasksIDs{
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
