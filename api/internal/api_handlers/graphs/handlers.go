package graphs

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/liriquew/control_system/api/internal/api_handlers/groups"
	"github.com/liriquew/control_system/api/internal/entities"
	graphsclient "github.com/liriquew/control_system/api/internal/grpc/clients/graphs"
	tasksclient "github.com/liriquew/control_system/api/internal/grpc/clients/tasks"
	jsontools "github.com/liriquew/control_system/api/internal/lib/json_tools"
	"github.com/liriquew/control_system/api/pkg/logger/sl"

	"github.com/liriquew/control_system/api/internal/models"
)

type GpraphsAPI interface {
	CreateGroupGraph(w http.ResponseWriter, r *http.Request)
	ListGroupGraphs(w http.ResponseWriter, r *http.Request)
	GetGraph(w http.ResponseWriter, r *http.Request)
	CreateNode(w http.ResponseWriter, r *http.Request)
	GetNode(w http.ResponseWriter, r *http.Request)
	UpdateNode(w http.ResponseWriter, r *http.Request)
	RemoveNode(w http.ResponseWriter, r *http.Request)
	GetDependencies(w http.ResponseWriter, r *http.Request)
	AddDependency(w http.ResponseWriter, r *http.Request)
	RemoveDependency(w http.ResponseWriter, r *http.Request)
	PredictGraph(w http.ResponseWriter, r *http.Request)
}

type Graphs struct {
	graphsClient graphsclient.GraphClient
	tasksClient  tasksclient.TasksClient
	log          *slog.Logger
}

func NewGraphsService(
	log *slog.Logger,
	graphsClient graphsclient.GraphClient,
	tasksClient tasksclient.TasksClient,
) *Graphs {
	return &Graphs{
		graphsClient: graphsClient,
		tasksClient:  tasksClient,
		log:          log,
	}
}

func (g *Graphs) CreateGroupGraph(w http.ResponseWriter, r *http.Request) {
	groupID := groups.GetGroupID(r)
	graph, err := entities.GraphWithNodesFromJSON(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	graph.GraphInfo.GroupID = groupID

	for _, node := range graph.Nodes {
		if node.TaskID == 0 {
			http.Error(w, fmt.Sprintf("taskID required: nodeID %d", node.ID), http.StatusBadRequest)
			return
		}
		if err := g.tasksClient.TaskExists(r.Context(), node.TaskID, groupID); err != nil {
			if errors.Is(err, tasksclient.ErrNotFound) {
				http.Error(w, fmt.Sprintf("taskID not found: taskID %d", node.TaskID), http.StatusNotFound)
				return
			}

			g.log.Error("error while checking is tasks in group", slog.Any("node", node), sl.Err(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	graphID, err := g.graphsClient.CreateGroupGraph(r.Context(), graph)
	if err != nil {
		if errors.Is(err, graphsclient.ErrBadParams) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		g.log.Error("error while creating group graph", slog.Any("graph", graph), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WriteInt64ID(w, graphID)
}

func (g *Graphs) ListGroupGraphs(w http.ResponseWriter, r *http.Request) {
	groupID := groups.GetGroupID(r)
	offset := GetOffset(r)

	graphs, err := g.graphsClient.ListGroupGraphs(r.Context(), groupID, offset)
	if err != nil {
		g.log.Error("error while listing group graphs", slog.Int64("groupID", groupID), slog.Int64("offset", offset), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, graphs)
}

func (g *Graphs) GetGraph(w http.ResponseWriter, r *http.Request) {
	graphID := GetGraphID(r)

	graph, err := g.graphsClient.GetGraph(r.Context(), graphID)
	if err != nil {
		if errors.Is(err, graphsclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, graph)
}

func (g *Graphs) CreateNode(w http.ResponseWriter, r *http.Request) {
	graphID := r.Context().Value(GraphID{}).(int64)

	node, err := models.NodeModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}
	node.GraphID = graphID

	groupID := groups.GetGroupID(r)
	if err := g.tasksClient.TaskExists(r.Context(), node.TaskID, groupID); err != nil {
		if errors.Is(err, tasksclient.ErrNotFound) {
			http.Error(w, "task not found in group", http.StatusNotFound)
			return
		}

		g.log.Error("error while checking is task exists in group", slog.Any("node", node), slog.Int64("groupID", groupID), sl.Err(err))
		return
	}

	nodeID, err := g.graphsClient.CreateNode(r.Context(), node)
	if err != nil {
		if errors.Is(err, graphsclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, graphsclient.ErrBadParams) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		g.log.Error("error while creating node", slog.Any("node", node), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WriteInt64ID(w, nodeID)
}

func (g *Graphs) GetNode(w http.ResponseWriter, r *http.Request) {
	nodeID := GetNodeID(r)

	node, err := g.graphsClient.GetNode(r.Context(), nodeID)
	if err != nil {
		if errors.Is(err, graphsclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		g.log.Error("error while getting node", slog.Int64("nodeID", nodeID), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, node)
}

func (g *Graphs) UpdateNode(w http.ResponseWriter, r *http.Request) {
	node, err := models.NodeModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusInternalServerError)
		return
	}
	node.ID = GetNodeID(r)
	node.GraphID = GetGraphID(r)

	err = g.graphsClient.UpdateNode(r.Context(), node)
	if err != nil {
		if errors.Is(err, graphsclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, graphsclient.ErrBadParams) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		g.log.Error("error while updating node", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Graphs) RemoveNode(w http.ResponseWriter, r *http.Request) {
	graphID := GetGraphID(r)
	nodeID := GetNodeID(r)

	err := g.graphsClient.RemoveNode(r.Context(), graphID, nodeID)
	if err != nil {
		if errors.Is(err, graphsclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		g.log.Error("error while removing node", slog.Int64("graphID", graphID), slog.Int64("nodeID", nodeID), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Graphs) GetDependencies(w http.ResponseWriter, r *http.Request) {
	graphID := r.Context().Value(GraphID{}).(int64)
	nodeID := r.Context().Value(NodeID{}).(int64)

	deps, err := g.graphsClient.GetDependencies(r.Context(), graphID, nodeID)
	if err != nil {
		if errors.Is(err, graphsclient.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		g.log.Error("error while getting dependencies", slog.Int64("graphID", graphID), slog.Int64("nodeID", nodeID), sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, deps)
}

func (g *Graphs) AddDependency(w http.ResponseWriter, r *http.Request) {
	graphID := GetGraphID(r)
	nodeID := GetNodeID(r)
	dependencyNodeID := GetDependencyNodeID(r)

	dependency := &models.Dependency{
		FromNodeID: nodeID,
		ToNodeID:   dependencyNodeID,
		GraphID:    graphID,
	}

	err := g.graphsClient.AddDependency(r.Context(), dependency)
	if err != nil {
		if errors.Is(err, graphsclient.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, graphsclient.ErrBadParams) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, graphsclient.ErrAlreadyExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}

		g.log.Error("error while adding dependency", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Graphs) RemoveDependency(w http.ResponseWriter, r *http.Request) {
	graphID := r.Context().Value(GraphID{}).(int64)
	nodeID := r.Context().Value(NodeID{}).(int64)
	dependencyNodeID := r.Context().Value(DependencyNodeID{}).(int64)

	dependency := &models.Dependency{
		FromNodeID: nodeID,
		ToNodeID:   dependencyNodeID,
		GraphID:    graphID,
	}

	err := g.graphsClient.RemoveDependency(r.Context(), dependency)
	if err != nil {
		if errors.Is(err, graphsclient.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
		if errors.Is(err, graphsclient.ErrBadParams) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		g.log.Error("error while removing dependency", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Graphs) PredictGraph(w http.ResponseWriter, r *http.Request) {
	graphID := r.Context().Value(GraphID{}).(int64)
	priority := r.Context().Value(Priority{}).(int64)

	predictedGraph, err := g.graphsClient.PredictGraph(r.Context(), graphID, int(priority))
	if err != nil {
		if errors.Is(err, graphsclient.ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if errors.Is(err, graphsclient.ErrBadGraph) {
			http.Error(w, err.Error(), http.StatusPreconditionFailed)
			return
		}

		g.log.Error("error while predicting graph", sl.Err(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsontools.WtiteJSON(w, predictedGraph)
}
