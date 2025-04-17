package predictionsclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	"github.com/liriquew/control_system/internal/lib/config"
	prdt_pb "github.com/liriquew/control_system/services_protos/predictions_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
)

type PredicionsClient struct {
	client prdt_pb.PredictionsClient
	log    *slog.Logger
}

func NewPredictionsClient(log *slog.Logger, cfg config.ClientConfig) (*PredicionsClient, error) {
	const op = "predictionsclient.New"

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

	return &PredicionsClient{
		client: prdt_pb.NewPredictionsClient(cc),
		log:    log,
	}, nil
}

func InterceptorLogger(log *slog.Logger) grpclog.Logger {
	return grpclog.LoggerFunc(func(ctx context.Context, level grpclog.Level, msg string, fields ...any) {
		log.Log(ctx, slog.Level(level), msg, fields...)
	})
}

var (
	ErrNotFound = errors.New("not found")
	ErrInternal = errors.New("internal")
)

// func (c *PredicionsClient) Predict(ctx context.Context, task *models.Task) (float64, error) {
// 	var userID int64
// 	if task.GroupID != 0 {
// 		userID = task.AssignedTo
// 	} else {
// 		userID = task.CreatedBy
// 	}
// 	predicted, err := c.client.Predict(ctx, &prdt_pb.PredictRequest{
// 		UID:         userID,
// 		PlannedTime: task.PlannedTime,
// 	})
// 	if err != nil {
// 		return 0, err
// 	}

// 	return predicted.ActualTime, nil
// }
