package grphclient

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	grpclog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/retry"

	"github.com/liriquew/control_system/api/internal/entities"
	"github.com/liriquew/control_system/api/internal/lib/config"
	"github.com/liriquew/control_system/api/internal/lib/converter"
	"github.com/liriquew/control_system/api/internal/models"
	grph_pb "github.com/liriquew/control_system/services_protos/graphs_service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type GRPCGraphClient struct {
	client grph_pb.GraphsClient
	log    *slog.Logger
}

func NewGraphClient(log *slog.Logger, cfg config.ClientConfig) (*GRPCGraphClient, error) {
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

	return &GRPCGraphClient{
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
	ErrTaskInNode    = errors.New("task in node")
	ErrBadParams     = errors.New("")
	ErrNotFound      = errors.New("")
	ErrBadGraph      = errors.New("")
	ErrAlreadyExists = errors.New("")
)

type GraphClient interface {
	CreateGroupGraph(ctx context.Context, graph *entities.GraphWithNodes) (int64, error)
	ListGroupGraphs(ctx context.Context, groupID, offset int64) ([]*entities.GraphWithNodes, error)
	GetGraph(ctx context.Context, graphID int64) (*entities.GraphWithNodes, error)
	GetNode(ctx context.Context, graphID int64) (*models.Node, error)
	CreateNode(ctx context.Context, node *models.Node) (int64, error)
	UpdateNode(ctx context.Context, node *models.Node) error
	RemoveNode(ctx context.Context, graphID, nodeID int64) error
	GetDependencies(ctx context.Context, graphID, nodeID int64) (*models.Node, error)
	AddDependency(ctx context.Context, dep *models.Dependency) error
	RemoveDependency(ctx context.Context, dep *models.Dependency) error
	PredictGraph(ctx context.Context, graphID int64) (*entities.PredictedGraph, error)
	TaskInNode(ctx context.Context, taskID int64) (int64, error)
}

func (c *GRPCGraphClient) CreateGroupGraph(ctx context.Context, graph *entities.GraphWithNodes) (int64, error) {
	resp, err := c.client.CreateGroupGraph(ctx, converter.ConvertGraphWithNodesToProto(graph))
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.InvalidArgument:
				return 0, fmt.Errorf("%w:%s", ErrBadParams, st.Message())
			}
		}

		return 0, err
	}

	return resp.GraphID, nil
}

func (c *GRPCGraphClient) ListGroupGraphs(ctx context.Context, groupID, offset int64) ([]*entities.GraphWithNodes, error) {
	resp, err := c.client.ListGroupGraphs(ctx, &grph_pb.ListGroupGraphsRequest{
		GroupID: groupID,
		Offset:  offset,
	})
	if err != nil {
		return nil, err
	}

	return converter.ConvertGraphsWithNodesToModel(resp.Graphs), nil
}

func (c *GRPCGraphClient) GetGraph(ctx context.Context, graphID int64) (*entities.GraphWithNodes, error) {
	resp, err := c.client.GetGraph(ctx, &grph_pb.GetGraphRequest{
		GraphID: graphID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, ErrNotFound
			}
		}
		return nil, err
	}

	return converter.ConvertGraphWithNodesToModel(resp), nil
}

func (c *GRPCGraphClient) CreateNode(ctx context.Context, node *models.Node) (int64, error) {
	resp, err := c.client.CreateNode(ctx, &grph_pb.CreateNodeRequest{
		Node: converter.ConvertNodeToProto(node),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return 0, ErrNotFound
			case codes.InvalidArgument:
				return 0, fmt.Errorf("%w%s", ErrBadParams, st.Message())
			}
		}
		return 0, err
	}

	return resp.Node.ID, nil
}

func (c *GRPCGraphClient) GetNode(ctx context.Context, nodeID int64) (*models.Node, error) {
	resp, err := c.client.GetNode(ctx, &grph_pb.GetNodeRequest{
		NodeID: nodeID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, ErrNotFound
			}
		}
		return nil, err
	}

	return converter.ConvertNodeToModel(resp.Node), nil
}

func (c *GRPCGraphClient) UpdateNode(ctx context.Context, node *models.Node) error {
	_, err := c.client.UpdateNode(ctx, &grph_pb.UpdateNodeRequest{
		Node: converter.ConvertNodeToProto(node),
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return ErrNotFound
			case codes.InvalidArgument:
				return fmt.Errorf("%w%s", ErrBadParams, st.Message())
			}
		}
		return err
	}

	return nil
}

func (c *GRPCGraphClient) RemoveNode(ctx context.Context, graphID, nodeID int64) error {
	_, err := c.client.RemoveNode(ctx, &grph_pb.RemoveNodeRequest{
		GraphID: graphID,
		NodeID:  nodeID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return ErrNotFound
			}
		}
		return err
	}

	return nil
}

func (c *GRPCGraphClient) GetDependencies(ctx context.Context, graphID, nodeID int64) (*models.Node, error) {
	resp, err := c.client.GetDependencies(ctx, &grph_pb.GetDependenciesRequest{
		GraphID: graphID,
		NodeID:  nodeID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, ErrNotFound
			}
		}
		return nil, err
	}

	return converter.ConvertNodeToModel(resp.Node), nil
}

func (c *GRPCGraphClient) AddDependency(ctx context.Context, dep *models.Dependency) error {
	_, err := c.client.AddDependency(ctx, &grph_pb.DependencyRequest{
		GraphID: dep.GraphID,
		Dependency: &grph_pb.Dependency{
			FromNodeID: dep.FromNodeID,
			ToNodeID:   dep.ToNodeID,
		},
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return fmt.Errorf("%w%s", ErrNotFound, st.Message())
			case codes.InvalidArgument:
				return fmt.Errorf("%w%s", ErrBadParams, st.Message())
			case codes.AlreadyExists:
				return ErrAlreadyExists
			}
		}
		return err
	}

	return nil
}

func (c *GRPCGraphClient) RemoveDependency(ctx context.Context, dep *models.Dependency) error {
	_, err := c.client.RemoveDependency(ctx, &grph_pb.DependencyRequest{
		GraphID: dep.GraphID,
		Dependency: &grph_pb.Dependency{
			FromNodeID: dep.FromNodeID,
			ToNodeID:   dep.ToNodeID,
		},
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return fmt.Errorf("%w%s", ErrNotFound, st.Message())
			case codes.InvalidArgument:
				return fmt.Errorf("%w%s", ErrBadParams, st.Message())
			}
		}
		return err
	}

	return nil
}

func (c *GRPCGraphClient) PredictGraph(ctx context.Context, graphID int64) (*entities.PredictedGraph, error) {
	resp, err := c.client.PredictGraph(ctx, &grph_pb.PredictGraphRequest{
		GraphID: graphID,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, fmt.Errorf("%w%s", ErrNotFound, st.Message())
			case codes.FailedPrecondition:
				return nil, fmt.Errorf("%w%s", ErrBadGraph, st.Message())
			}
		}
		return nil, err
	}

	return converter.ConvertPredictedGraph(resp), nil
}

func (c *GRPCGraphClient) TaskInNode(ctx context.Context, taskID int64) (int64, error) {
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
