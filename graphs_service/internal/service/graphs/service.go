package graphsservice

import (
	"context"
	"errors"
	"log/slog"
	"strconv"

	"github.com/liriquew/control_system/graphs_service/internal/entities"
	tasksclient "github.com/liriquew/control_system/graphs_service/internal/grpc/clients/tasks"
	graphtools "github.com/liriquew/control_system/graphs_service/internal/lib/graph_tools"
	graph_wrapper "github.com/liriquew/control_system/graphs_service/internal/lib/graph_tools/wrapper"
	"github.com/liriquew/control_system/graphs_service/internal/models"
	"github.com/liriquew/control_system/graphs_service/internal/repository"
	"github.com/liriquew/control_system/graphs_service/pkg/logger/sl"
	grph_pb "github.com/liriquew/control_system/services_protos/graphs_service"
	tsks_pb "github.com/liriquew/control_system/services_protos/tasks_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Repository interface {
	GetGraphGroup(ctx context.Context, graphID int64) (int64, error)

	CreateGraph(ctx context.Context, graph *grph_pb.Graph, nodes []*grph_pb.Node) (int64, error)
	ListGroupGraphs(ctx context.Context, groupID, offset int64) ([]*entities.GraphWithNodes, error)

	GetGraph(ctx context.Context, graphID int64) (*entities.GraphWithNodes, error)
	CreateNode(ctx context.Context, node *grph_pb.Node) (int64, error)
	GetNode(ctx context.Context, graphID, nodeID int64) (*models.Node, error)
	UpdateNode(ctx context.Context, node *grph_pb.Node) error
	RemoveNode(ctx context.Context, nodeID int64) error
	GetDependencies(ctx context.Context, nodeID int64) (*models.Node, error)
	AddDependency(ctx context.Context, graphID int64, dep *grph_pb.Dependency) error
	RemoveDependensy(ctx context.Context, dependency *grph_pb.Dependency) error

	TaskInNode(ctx context.Context, taskID int64) (int64, error)
}

type tasksClient interface {
	GetPredictedTasks(context.Context, []int64) ([]*tsks_pb.PredictedTask, []int64, error)
	TaskExists(context.Context, int64, int64) error
}

type authClient interface {
	Authenticate(context.Context, string) (int64, error)
}

type Service struct {
	grph_pb.UnimplementedGraphsServer
	repository  Repository
	log         *slog.Logger
	tasksClient tasksClient
}

func New(log *slog.Logger, graphsRepository Repository, tc tasksClient) *Service {
	return &Service{
		log:         log,
		repository:  graphsRepository,
		tasksClient: tc,
	}
}

func (s *Service) authenticate(ctx context.Context) (int64, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.log.Error("error while extracting metadata")
		return 0, status.Error(codes.Unauthenticated, "missing metadata")
	}

	AuthParams := md.Get("user-id")
	if len(AuthParams) == 0 {
		return 0, status.Error(codes.Unauthenticated, "missing user-idmetadata")
	}
	userID, err := strconv.ParseInt(AuthParams[0], 10, 64)
	if err != nil {
		return 0, status.Error(codes.Unauthenticated, "invalid user-id metadata")
	}

	return userID, nil
}

func (s *Service) CreateGroupGraph(ctx context.Context, req *grph_pb.GraphWithNodes) (*grph_pb.GraphResponse, error) {
	userID, err := s.authenticate(ctx)
	if err != nil {
		s.log.Error("error while authenticate user", sl.Err(err))
		return nil, err
	}
	req.GraphInfo.CreatedBy = userID

	if req.GraphInfo.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "empty graph name")
	}

	graphID, err := s.repository.CreateGraph(ctx, req.GraphInfo, req.Nodes)
	if err != nil {
		s.log.Debug("error while creating graph", sl.Err(err))
		if errors.Is(err, repository.ErrTaskAlreadyInNode) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, repository.ErrDependencyNotFound) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		s.log.Error("error while creating graph ", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &grph_pb.GraphResponse{
		GraphID: graphID,
	}, nil
}

func (s *Service) ListGroupGraphs(ctx context.Context, req *grph_pb.ListGroupGraphsRequest) (*grph_pb.GraphListResponse, error) {
	graphs, err := s.repository.ListGroupGraphs(ctx, req.GroupID, req.Offset)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		s.log.Error("error while getting user's graphs", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	resp := make([]*grph_pb.GraphWithNodes, 0, len(graphs))
	for _, graph := range graphs {
		resp = append(resp, &grph_pb.GraphWithNodes{
			GraphInfo: models.ConvertGraphToProto(graph.GraphInfo),
			Nodes:     models.ConvertNodesToProto(graph.Nodes),
		})
	}

	return &grph_pb.GraphListResponse{
		Graphs: resp,
	}, nil
}

func (s *Service) GetGraph(ctx context.Context, req *grph_pb.GetGraphRequest) (*grph_pb.GraphWithNodes, error) {
	graph, err := s.repository.GetGraph(ctx, req.GraphID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "graph not found")
		}

		s.log.Error("error while getting graph", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &grph_pb.GraphWithNodes{
		GraphInfo: models.ConvertGraphToProto(graph.GraphInfo),
		Nodes:     models.ConvertNodesToProto(graph.Nodes),
	}, nil
}

func (s *Service) GetNode(ctx context.Context, req *grph_pb.GetNodeRequest) (*grph_pb.NodeResponse, error) {
	node, err := s.repository.GetNode(ctx, req.GraphID, req.NodeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "node not found")
		}

		s.log.Error("error while getting node", sl.Err(err))
		return nil, status.Error(codes.Internal, "internals")
	}

	return &grph_pb.NodeResponse{
		Node: models.ConvertNodeToProto(node),
	}, nil
}

func (s *Service) CreateNode(ctx context.Context, req *grph_pb.CreateNodeRequest) (*grph_pb.NodeResponse, error) {
	groupID, err := s.repository.GetGraphGroup(ctx, req.Node.GraphID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "graph not found")
		}

		s.log.Error("error while getting group in GetGraph", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if err := s.tasksClient.TaskExists(ctx, req.Node.TaskID, groupID); err != nil {
		if errors.Is(err, tasksclient.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}

		s.log.Error("error while checking is task exists", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	nodeID, err := s.repository.CreateNode(ctx, req.Node)
	if err != nil {
		if errors.Is(err, repository.ErrNotExists) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, repository.ErrSelfDependencyRejected) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, repository.ErrTaskAlreadyInNode) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		s.log.Error("error while creating node", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	req.Node.ID = nodeID

	return &grph_pb.NodeResponse{
		Node: req.Node,
	}, nil
}

func (s *Service) UpdateNode(ctx context.Context, req *grph_pb.UpdateNodeRequest) (*emptypb.Empty, error) {
	groupID, err := s.repository.GetGraphGroup(ctx, req.Node.GraphID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "graph not found")
		}

		s.log.Error("error while getting group in GetGraph", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if err := s.tasksClient.TaskExists(ctx, req.Node.TaskID, groupID); err != nil {
		if errors.Is(err, tasksclient.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "task not found")
		}

		s.log.Error("error while checking is task exists", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	if err := s.repository.UpdateNode(ctx, req.Node); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "node not found")
		}

		s.log.Error("error while updating node", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) RemoveNode(ctx context.Context, req *grph_pb.RemoveNodeRequest) (*emptypb.Empty, error) {
	if err := s.repository.RemoveNode(ctx, req.NodeID); err != nil && !errors.Is(err, repository.ErrNotExists) {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "node not found")
		}

		s.log.Error("error while removing node", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) GetDependencies(ctx context.Context, req *grph_pb.GetDependenciesRequest) (*grph_pb.NodeWithDependencies, error) {
	node, err := s.repository.GetDependencies(ctx, req.NodeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "node not found")
		}

		s.log.Error("error while removing node", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &grph_pb.NodeWithDependencies{
		Node: models.ConvertNodeToProto(node),
	}, nil
}

func (s *Service) AddDependency(ctx context.Context, req *grph_pb.DependencyRequest) (*emptypb.Empty, error) {
	if err := s.repository.AddDependency(ctx, req.GraphID, req.Dependency); err != nil {
		if errors.Is(err, repository.ErrNotExists) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		if errors.Is(err, repository.ErrSelfDependencyRejected) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}

		s.log.Error("error while removing node", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) RemoveDependency(ctx context.Context, req *grph_pb.DependencyRequest) (*emptypb.Empty, error) {
	if err := s.repository.RemoveDependensy(ctx, req.Dependency); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "node not found")
		}

		s.log.Error("error while removing node", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) PredictGraph(ctx context.Context, req *grph_pb.PredictGraphRequest) (*grph_pb.PredictedGraphResponse, error) {
	graph, err := s.repository.GetGraph(ctx, req.GraphID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "graph not found")
		}

		s.log.Error("error while getting graph in Predict", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	tasksIDs := make([]int64, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		tasksIDs = append(tasksIDs, node.TaskID)
	}

	predictedTasks, unpredictedUIDs, err := s.tasksClient.GetPredictedTasks(ctx, tasksIDs)
	if err != nil {
		if errors.Is(err, tasksclient.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "tasks not found")
		}

		s.log.Error("error while getting tasks in Predict", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	tasksNodesMap := make(map[int64]*grph_pb.Node, len(predictedTasks))
	for _, node := range graph.Nodes {
		tasksNodesMap[node.TaskID] = models.ConvertNodeToProto(node)
	}

	nodesWithTasks := make([]*grph_pb.NodeWithTask, 0, len(predictedTasks))
	for _, predictedTask := range predictedTasks {
		nodesWithTasks = append(nodesWithTasks, &grph_pb.NodeWithTask{
			Node:          tasksNodesMap[predictedTask.Task.ID],
			Task:          predictedTask.Task,
			PredictedTime: predictedTask.PredictedTime,
		})
	}

	predictableGraph := &grph_pb.PredictedGraphResponse{
		Graph:           models.ConvertGraphToProto(graph.GraphInfo),
		Nodes:           nodesWithTasks,
		UnpredictedUIDs: unpredictedUIDs,
	}

	nodesValuesMap := make(map[int64]float64, len(predictableGraph.Nodes))
	for _, node := range predictableGraph.Nodes {
		nodesValuesMap[node.Node.ID] = node.PredictedTime
	}

	paths, err := graphtools.FindCriticalPath(graph_wrapper.WrapPredictedGraph(predictableGraph), nodesValuesMap)
	if err != nil {
		if errors.Is(err, graphtools.ErrCycleInGraph) {
			return nil, status.Error(codes.FailedPrecondition, "cycle found in graph")
		}

		s.log.Error("error while searching critical path", sl.Err(err))
		return nil, err
	}

	for _, path := range paths {
		predictableGraph.Paths = append(predictableGraph.Paths, &grph_pb.Path{
			NodeIDs: path,
		})
	}

	return predictableGraph, nil
}

func (s *Service) TaskInNode(ctx context.Context, req *grph_pb.TaskInNodeRequest) (*grph_pb.TaskInNodeResponse, error) {
	nodeID, err := s.repository.TaskInNode(ctx, req.TaskID)
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		s.log.Error("error while searching task in node", sl.Err(err))
		return nil, status.Error(codes.Internal, "internal")
	}

	return &grph_pb.TaskInNodeResponse{NodeID: nodeID}, nil
}
