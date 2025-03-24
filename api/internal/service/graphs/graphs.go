package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/liriquew/control_system/internal/entities"
	predictions_client "github.com/liriquew/control_system/internal/grpc/client"
	graphtools "github.com/liriquew/control_system/internal/lib/graph_tools"
	"github.com/liriquew/control_system/internal/models"
	graphs_tasks_repository "github.com/liriquew/control_system/internal/repository/graph_tasks"
	repository "github.com/liriquew/control_system/internal/repository/graphs"
)

type GraphsRepository interface {
	CheckEditorPermission(ctx context.Context, userID, graphID int64) error
	CheckUserPermission(ctx context.Context, userID, graphID int64) error

	GetGraph(ctx context.Context, graphID int64) (*entities.GraphWithNodes, error)
	CreateNode(ctx context.Context, userID, graphID int64, node *models.Node) (*models.Node, error)
	GetNode(ctx context.Context, graphID, nodeID int64) (*models.Node, error)
	UpdateNode(ctx context.Context, userID, graphID int64, node *models.Node) error
	RemoveNode(ctx context.Context, userID, graphID, nodeID int64) error
	GetDependencies(ctx context.Context, userID, graphID, nodeID int64) (*models.Node, error)
	AddDependency(ctx context.Context, userID, graphID int64, node *models.Dependency) (*models.Dependency, error)
	RemoveDependensy(ctx context.Context, userID, graphID int64, dependency *models.Dependency) error
}

type GraphsTasksRepository interface {
	GetTasksFromNodes(ctx context.Context, nodes []*models.Node) ([]*models.Task, error)
}

type GraphsService struct {
	graphsRepo      GraphsRepository
	graphsTasksRepo GraphsTasksRepository
	predictorClient *predictions_client.Client
	infoLog         *log.Logger
	errorLog        *log.Logger
}

func NewGraphsService(repo GraphsRepository, predictionsClient *predictions_client.Client, graph_tasks_repo GraphsTasksRepository, infoLog, errorLog *log.Logger) (*GraphsService, error) {
	return &GraphsService{
		graphsRepo:      repo,
		graphsTasksRepo: graph_tasks_repo,
		predictorClient: predictionsClient,
		infoLog:         infoLog,
		errorLog:        errorLog,
	}, nil
}

var (
	ErrDenied = fmt.Errorf("access denyed")

	ErrNotFound  = errors.New("not found")
	ErrNotExists = errors.New("")

	ErrNodeNotFound           = errors.New("node not found")
	ErrDepNotFound            = errors.New("dependency not found")
	ErrNothingToUpdate        = errors.New("empty updatable fields")
	ErrBadIDParams            = errors.New("bad ID params")
	ErrSelfDependencyRejected = errors.New("")

	ErrCycleInGraph      = errors.New("cycle found")
	ErrSomeTasksNotFound = errors.New("some tasks not found")
)

func (gs *GraphsService) checkEditorPermission(ctx context.Context, userID, graphID int64) error {
	if userID <= 0 || graphID <= 0 {
		return ErrBadIDParams
	}

	if err := gs.graphsRepo.CheckEditorPermission(ctx, userID, graphID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return ErrDenied
		}
		gs.errorLog.Println(err)
		return err
	}

	return nil
}

func (gs *GraphsService) checkUserPermission(ctx context.Context, userID, graphID int64) error {
	if userID <= 0 || graphID <= 0 {
		return ErrBadIDParams
	}

	if err := gs.graphsRepo.CheckUserPermission(ctx, userID, graphID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return ErrDenied
		}
		gs.errorLog.Println(err)
		return err
	}

	return nil
}

func (gs *GraphsService) GetGraph(ctx context.Context, userID, graphID int64) (*entities.GraphWithNodes, error) {
	if err := gs.checkUserPermission(ctx, userID, graphID); err != nil {
		return nil, err
	}

	graph, err := gs.graphsRepo.GetGraph(ctx, graphID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		gs.errorLog.Println(err)
		return nil, err
	}

	return graph, nil
}

func (gs *GraphsService) GetNode(ctx context.Context, userID, graphID, nodeID int64) (*models.Node, error) {
	if err := gs.graphsRepo.CheckUserPermission(ctx, userID, graphID); err != nil {
		if errors.Is(err, repository.ErrDenied) {
			return nil, ErrDenied
		}

		gs.errorLog.Println(err)
		return nil, fmt.Errorf("service: error while checkig user permissions: %w", err)
	}

	node, err := gs.graphsRepo.GetNode(ctx, graphID, nodeID)
	if err != nil {
		if errors.Is(err, repository.ErrNodeNotFound) {
			return nil, ErrNodeNotFound
		}

		gs.errorLog.Println(err)
		return nil, fmt.Errorf("service: error while getting node: %w", err)
	}

	return node, nil
}

func (gs *GraphsService) CreateNode(ctx context.Context, userID, graphID int64, node *models.Node) (*models.Node, error) {
	if err := gs.checkEditorPermission(ctx, userID, graphID); err != nil {
		return nil, err
	}

	node, err := gs.graphsRepo.CreateNode(ctx, userID, graphID, node)
	if err != nil {
		if errors.Is(err, repository.ErrNotExists) {
			return nil, fmt.Errorf("%w%w", err, ErrNotExists)
		}
		if errors.Is(err, repository.ErrSelfDependencyRejected) {
			return nil, fmt.Errorf("%w%w", err, ErrSelfDependencyRejected)
		}
		gs.errorLog.Println(err)
		return nil, err
	}

	return node, err
}

func (gs *GraphsService) RemoveNode(ctx context.Context, userID, graphID, nodeID int64) error {
	if err := gs.checkEditorPermission(ctx, userID, graphID); err != nil {
		return err
	}

	err := gs.graphsRepo.RemoveNode(ctx, userID, graphID, nodeID)
	if err != nil {
		if errors.Is(err, repository.ErrNotExists) {
			return ErrNotExists
		}

		gs.errorLog.Println(err)
		return err
	}

	return nil
}

func (gs *GraphsService) UpdateNode(ctx context.Context, userID, graphID int64, node *models.Node) error {
	if err := gs.checkEditorPermission(ctx, userID, graphID); err != nil {
		return err
	}

	err := gs.graphsRepo.UpdateNode(ctx, userID, graphID, node)
	if err != nil {
		if errors.Is(err, repository.ErrNothingToUpdate) {
			return ErrNothingToUpdate
		}
		if errors.Is(err, repository.ErrNotExists) {
			return ErrNotExists
		}
		gs.errorLog.Println(err)
		return err
	}

	return nil
}

func (gs *GraphsService) GetDependencies(ctx context.Context, userID, graphID, nodeID int64) (*models.Node, error) {
	if err := gs.checkUserPermission(ctx, userID, graphID); err != nil {
		return nil, err
	}

	deps, err := gs.graphsRepo.GetDependencies(ctx, userID, graphID, nodeID)
	if err != nil {
		if errors.Is(err, repository.ErrDependencyNotFound) {
			return nil, ErrDepNotFound
		}
		if errors.Is(err, repository.ErrNodeNotFound) {
			return nil, ErrNodeNotFound
		}
		gs.errorLog.Println(err)
		return nil, err
	}

	return deps, nil
}

func (gs *GraphsService) AddDependency(ctx context.Context, userID, graphID int64, dependency *models.Dependency) (*models.Dependency, error) {
	if err := gs.checkEditorPermission(ctx, userID, graphID); err != nil {
		return nil, err
	}

	dependency, err := gs.graphsRepo.AddDependency(ctx, userID, graphID, dependency)
	if err != nil {
		if errors.Is(err, repository.ErrNotExists) {
			return nil, fmt.Errorf("%s%w", err.Error(), ErrNotExists)
		}
		if errors.Is(err, repository.ErrSelfDependencyRejected) {
			return nil, ErrSelfDependencyRejected
		}
		gs.errorLog.Println(err)
		return nil, err
	}

	return dependency, nil
}

func (gs *GraphsService) RemoveDependensy(ctx context.Context, userID, graphID int64, dependency *models.Dependency) error {
	if err := gs.checkEditorPermission(ctx, userID, graphID); err != nil {
		return err
	}

	err := gs.graphsRepo.RemoveDependensy(ctx, userID, graphID, dependency)
	if err != nil {
		if errors.Is(err, repository.ErrNotExists) {
			return ErrNotExists
		}
		gs.errorLog.Println(err)
		return err
	}

	return nil
}

func (gs *GraphsService) PredictGraph(ctx context.Context, userID, graphID int64) (*entities.PredictedGraph, error) {
	if err := gs.checkUserPermission(ctx, userID, graphID); err != nil {
		return nil, err
	}

	graphWithNodes, err := gs.graphsRepo.GetGraph(ctx, graphID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		gs.errorLog.Println(err)
		return nil, err
	}

	nodesTasks, err := gs.graphsTasksRepo.GetTasksFromNodes(ctx, graphWithNodes.Nodes)
	if err != nil {
		if errors.Is(err, graphs_tasks_repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		gs.errorLog.Println(err)
		return nil, err
	}

	if len(nodesTasks) != len(graphWithNodes.Nodes) {
		return nil, ErrSomeTasksNotFound
	}

	tasksNodesMap := make(map[int64]*models.Node, len(nodesTasks))
	for _, node := range graphWithNodes.Nodes {
		tasksNodesMap[node.TaskID] = node
	}

	nodesWithTasks := make([]*entities.NodeWithTask, 0, len(nodesTasks))
	for _, task := range nodesTasks {
		nodesWithTasks = append(nodesWithTasks, &entities.NodeWithTask{
			Node: tasksNodesMap[task.ID],
			Task: task,
		})
	}

	predictedNodes, err := gs.predictorClient.PredictList(ctx, nodesWithTasks)
	if err != nil {
		return nil, err
	}

	predictableGraph := &entities.GraphWithTasks{
		GraphInfo:       graphWithNodes.GraphInfo,
		Nodes:           predictedNodes.Nodes,
		UnpredictedUIDs: predictedNodes.UnpredictedUIDs,
	}

	nodesValuesMap := make(map[int64]float64, len(predictableGraph.Nodes))
	for _, node := range predictableGraph.Nodes {
		nodesValuesMap[node.Node.ID] = node.PredictedTime
	}

	paths, err := graphtools.FindCriticalPath(predictableGraph, nodesValuesMap)
	if err != nil {
		if errors.Is(err, graphtools.ErrCycleInGraph) {
			return nil, ErrCycleInGraph
		}
		return nil, err
	}

	// predictable graph contains some additional edges!
	return &entities.PredictedGraph{
		Graph: predictableGraph,
		Paths: paths,
	}, nil
}
