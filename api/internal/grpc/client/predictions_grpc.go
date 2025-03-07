package predictions_client

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"time_manage/internal/config"
	predictions "time_manage/internal/grpc/gen"

	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	"time_manage/internal/storage"
)

type Client struct {
	api predictions.PredictionsClient
	log *log.Logger
}

func New(log *log.Logger, cfg *config.ServiceConfig) (*Client, error) {
	retryOpts := []grpcretry.CallOption{
		grpcretry.WithCodes(codes.NotFound, codes.Aborted, codes.DeadlineExceeded),
		grpcretry.WithMax(uint(1)),
		grpcretry.WithPerRetryTimeout(time.Second),
	}

	connStr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	cc, err := grpc.NewClient(connStr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			// grpclog.UnaryClientInterceptor(InterceptorLogger(log), logOpts...),
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		api: predictions.NewPredictionsClient(cc),
		log: log,
	}, nil
}

var (
	ErrFailedPrecondition = fmt.Errorf("user does not have any completed tasks")
	ErrInvalidArgument    = fmt.Errorf("invalid argument")
)

func (c *Client) RecalculateAndSaveTask(ctx context.Context, uid int64, task *storage.Task) error {
	const op = "predictions_client.RecalculateAndSaveTask"

	_, err := c.api.RecalculateAndSaveTask(ctx, &predictions.RecalculateAndSaveTaskRequest{
		ID:          task.ID,
		UID:         uid,
		Title:       task.Title,
		Description: task.Description,
		PlannedTime: task.PlannedTime,
		ActualTime:  *task.ActualTime,
	})

	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return fmt.Errorf("%w: %s", ErrInvalidArgument, st.Message())
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (c *Client) Predict(ctx context.Context, uid int64, plannedTime float64) (float64, error) {
	const op = "predictions_client.Predict"

	resp, err := c.api.Predict(ctx, &predictions.PredictRequest{
		UID:         uid,
		PlannedTime: plannedTime,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.FailedPrecondition:
				return 0, fmt.Errorf("%w: %s", ErrFailedPrecondition, st.Message())
			case codes.InvalidArgument:
				return 0, fmt.Errorf("%w: %s", ErrInvalidArgument, st.Message())
			}
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return resp.ActualTime, nil
}

func (c *Client) Recalculate(ctx context.Context, uid int64) error {
	const op = "predictions_client.Recalculate"

	_, err := c.api.Recalculete(ctx, &predictions.RecalculateRequest{
		UID: uid,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.FailedPrecondition:
				return fmt.Errorf("%w: %s", ErrFailedPrecondition, st.Message())
			case codes.InvalidArgument:
				return fmt.Errorf("%w: %s", ErrInvalidArgument, st.Message())
			}
		}
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
