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

	"github.com/liriquew/control_system/internal/config"
	"github.com/liriquew/control_system/internal/entities"
	pb2 "github.com/liriquew/control_system/internal/grpc/gen/predictions_service"

	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"
)

type Client struct {
	api pb2.PredictionsClient
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
			grpcretry.UnaryClientInterceptor(retryOpts...),
		),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		api: pb2.NewPredictionsClient(cc),
		log: log,
	}, nil
}

var (
	ErrFailedPrecondition = fmt.Errorf("user does not have any completed tasks")
	ErrInvalidArgument    = fmt.Errorf("invalid argument")
)

func (c *Client) Predict(ctx context.Context, uid int64, plannedTime float64) (float64, error) {
	const op = "predictions_client.Predict"

	resp, err := c.api.Predict(ctx, &pb2.PredictRequest{
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

func (c *Client) PredictList(ctx context.Context, nodesWithTask []*entities.NodeWithTask) (*entities.PredictedNodes, error) {
	const op = "predictions_client.PredictList"

	PlannedUserTime := make([]*pb2.UserWithTime, len(nodesWithTask))
	for i, node := range nodesWithTask {
		fmt.Println(*node.Node.AssignedTo)
		PlannedUserTime[i] = &pb2.UserWithTime{
			UID:  *node.Node.AssignedTo,
			Time: node.Task.PlannedTime,
		}
	}

	resp, err := c.api.PredictList(ctx, &pb2.PredictListRequest{
		PlannedUserTime: PlannedUserTime,
	})
	if err != nil {
		return nil, err
	}

	for i, PredictedUserTime := range resp.PredictedUserTime {
		nodesWithTask[i].PredictedTime = PredictedUserTime.Time
	}

	return &entities.PredictedNodes{Nodes: nodesWithTask, UnpredictedUIDs: resp.UnpredictedUIDs}, nil
}
