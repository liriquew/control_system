package predictionsclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	prdt_pb "github.com/liriquew/control_system/services_protos/predictions_service"
	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	"github.com/liriquew/tasks_service/internal/lib/config"
	"github.com/liriquew/tasks_service/internal/models"
	"github.com/liriquew/tasks_service/pkg/logger/sl"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
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

func (c *PredicionsClient) Predict(ctx context.Context, task *models.Task) (float64, error) {
	var userID int64
	if task.GroupID.Int64 != 0 {
		userID = task.AssignedTo.Int64
	} else {
		userID = task.CreatedBy
	}
	predicted, err := c.client.Predict(ctx, &prdt_pb.PredictRequest{Info: &prdt_pb.PredictInfo{
		UID:         userID,
		TagsIDs:     task.Tags,
		PlannedTime: task.PlannedTime,
	}})
	if err != nil {
		return 0, err
	}

	return predicted.ActualTime, nil
}

func (c *PredicionsClient) PredictList(ctx context.Context, tasks []*models.Task) ([]*tsks_pb.PredictedTask, []int64, error) {
	timesToPredict := make([]*prdt_pb.PredictInfo, 0, len(tasks))
	taskTimesMap := make(map[int64]float64, len(tasks))
	for _, task := range tasks {
		if task.ActualTime != 0 {
			taskTimesMap[task.ID] = task.ActualTime
			continue
		}
		timesToPredict = append(timesToPredict, &prdt_pb.PredictInfo{
			ID:          task.ID,
			UID:         task.AssignedTo.Int64,
			PlannedTime: task.PlannedTime,
			TagsIDs:     task.Tags,
		})
	}

	c.log.Debug("user times", slog.Any("times", timesToPredict))
	var unpredictedUIDs []int64
	if len(timesToPredict) != 0 {
		predictedTimes := &prdt_pb.PredictListResponse{}
		predictedTimes, err := c.client.PredictList(ctx, &prdt_pb.PredictListRequest{
			Infos: timesToPredict,
		})
		if err != nil {
			if st, ok := status.FromError(err); ok {
				switch st.Code() {
				case codes.NotFound:
					return nil, nil, ErrNotFound
				}

				c.log.Error("error while predicting list of tasks", sl.Err(err))
				return nil, nil, ErrInternal
			}
		}
		unpredictedUIDs = predictedTimes.UnpredictedUIDs
		for _, predictedTime := range predictedTimes.PredictedUserTime {
			c.log.Debug("PREDICTED TIME:", slog.Int64("ID", predictedTime.ID), slog.Float64("TIME", predictedTime.PredictedTime))
			taskTimesMap[predictedTime.ID] = predictedTime.PredictedTime
		}
	}

	res := make([]*tsks_pb.PredictedTask, 0, len(tasks))
	for _, task := range tasks {
		res = append(res, &tsks_pb.PredictedTask{
			Task:          models.ConvertModelToProto(task),
			PredictedTime: taskTimesMap[task.ID],
		})
	}

	return res, unpredictedUIDs, nil
}
