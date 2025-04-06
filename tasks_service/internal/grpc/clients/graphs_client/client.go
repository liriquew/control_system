package grphclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	grph_pb "github.com/liriquew/control_system/services_protos/graphs_service"
	"github.com/liriquew/tasks_service/internal/lib/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type GraphClient struct {
	client grph_pb.GraphsClient
	log    *slog.Logger
}

func NewGraphClient(log *slog.Logger, cfg config.ClientConfig) (*GraphClient, error) {
	const op = "graphclient.New"

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

	return &GraphClient{
		client: grph_pb.NewGraphsClient(cc),
		log:    log,
	}, nil
}

func InterceptorLogger(log *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		log.Log(ctx, slog.Level(level), msg, fields...)
	})
}

var (
	ErrTaskInNode = errors.New("task in node")
)

func (c *GraphClient) TaskInNode(ctx context.Context, taskID int64) (int64, error) {
	nodeID, err := c.client.TaskInNode(ctx, &grph_pb.TaskInNodeRequest{
		TaskID: taskID,
	})
	if err != nil {
		return 0, err
	}

	if nodeID.NodeID != 0 {
		return nodeID.NodeID, fmt.Errorf("%w: %d", ErrTaskInNode, nodeID.NodeID)
	}

	return 0, nil
}
