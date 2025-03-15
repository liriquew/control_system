package graphs

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time_manage/internal/api_handlers/auth"
	"time_manage/internal/entities"
	jsontools "time_manage/internal/lib/json_tools"
	"time_manage/internal/models"
	service "time_manage/internal/service/graphs"
)

type GpraphsAPI interface {
	GetGraph(w http.ResponseWriter, r *http.Request)
	CreateNode(w http.ResponseWriter, r *http.Request)
	GetNode(w http.ResponseWriter, r *http.Request)
	UpdateNode(w http.ResponseWriter, r *http.Request)
	RemoveNode(w http.ResponseWriter, r *http.Request)
	GetDependencies(w http.ResponseWriter, r *http.Request)
	AddDependency(w http.ResponseWriter, r *http.Request)
	RemoveDependensy(w http.ResponseWriter, r *http.Request)
}

type GraphsServiceInterface interface {
	GetGraph(ctx context.Context, userID, graphID int64) (*entities.GraphWithNodes, error)
	CreateNode(ctx context.Context, userID, graphID int64, node *models.Node) (*models.Node, error)
	GetNode(ctx context.Context, userID, graphID, NodeID int64) (*models.Node, error)
	RemoveNode(ctx context.Context, userID, graphID, nodeID int64) error
	UpdateNode(ctx context.Context, userID, graphID int64, node *models.Node) error
	GetDependencies(ctx context.Context, userID, graphID, nodeID int64) (*models.Node, error)
	AddDependency(ctx context.Context, userID, graphID int64, dependency *models.Dependency) (*models.Dependency, error)
	RemoveDependensy(ctx context.Context, userID, graphID int64, dependency *models.Dependency) error
}

type Graphs struct {
	service  GraphsServiceInterface
	infoLog  *log.Logger
	errorLog *log.Logger
}

func New(service GraphsServiceInterface, infoLog, errorLog *log.Logger) *Graphs {
	return &Graphs{
		service:  service,
		infoLog:  infoLog,
		errorLog: errorLog,
	}
}

func (g *Graphs) GetGraph(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	graphID := r.Context().Value(GraphID{}).(int64)

	graph, err := g.service.GetGraph(r.Context(), userID, graphID)
	if err != nil {
		g.errorLog.Println("GetGraph", err)
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(graph)
	w.WriteHeader(http.StatusOK)
}

func (g *Graphs) CreateNode(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	graphID := r.Context().Value(GraphID{}).(int64)

	node, err := models.NodeModelFromJson(r.Body)
	if err != nil {
		g.errorLog.Println("bad json")
		http.Error(w, "bad json", http.StatusBadRequest)
		return
	}

	node, err = g.service.CreateNode(r.Context(), userID, graphID, node)
	if err != nil {
		g.errorLog.Println(err)
		if errors.Is(err, service.ErrNotExists) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrSelfDependencyRejected) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(node)
}

func (g *Graphs) GetNode(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	graphID := r.Context().Value(GraphID{}).(int64)
	nodeID := r.Context().Value(NodeID{}).(int64)

	node, err := g.service.GetNode(r.Context(), userID, graphID, nodeID)
	if err != nil {
		if errors.Is(err, service.ErrNodeNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	jsontools.WtiteJSON(w, node)
}

func (g *Graphs) RemoveNode(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	graphID := r.Context().Value(GraphID{}).(int64)
	nodeID := r.Context().Value(NodeID{}).(int64)

	err := g.service.RemoveNode(r.Context(), userID, graphID, nodeID)
	if err != nil {
		g.errorLog.Println(err)
		if errors.Is(err, service.ErrNotExists) {
			http.Error(w, "node not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Graphs) UpdateNode(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	graphID := r.Context().Value(GraphID{}).(int64)
	nodeID := r.Context().Value(NodeID{}).(int64)

	node, err := models.NodeModelFromJson(r.Body)
	if err != nil {
		http.Error(w, "bad json", http.StatusInternalServerError)
		return
	}
	node.ID = nodeID

	err = g.service.UpdateNode(r.Context(), userID, graphID, node)
	if err != nil {
		g.errorLog.Println(err)
		if errors.Is(err, service.ErrNothingToUpdate) {
			http.Error(w, "empty updateble fields", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrNotExists) {
			http.Error(w, "node does not exists", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Graphs) GetDependencies(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	graphID := r.Context().Value(GraphID{}).(int64)
	nodeID := r.Context().Value(NodeID{}).(int64)

	deps, err := g.service.GetDependencies(r.Context(), userID, graphID, nodeID)
	if err != nil {
		g.errorLog.Println(err)
		if errors.Is(err, service.ErrDepNotFound) {
			http.Error(w, "dependencies not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrNodeNotFound) {
			http.Error(w, "node not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	jsontools.WtiteJSON(w, deps)
}

func (g *Graphs) AddDependency(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	graphID := r.Context().Value(GraphID{}).(int64)
	nodeID := r.Context().Value(NodeID{}).(int64)
	dependencyNodeID := r.Context().Value(DependencyNodeID{}).(int64)

	dependency := &models.Dependency{
		FromNodeID: nodeID,
		ToNodeID:   dependencyNodeID,
	}

	_, err := g.service.AddDependency(r.Context(), userID, graphID, dependency)
	if err != nil {
		g.errorLog.Println(err)
		if errors.Is(err, service.ErrNotExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		if errors.Is(err, service.ErrSelfDependencyRejected) {
			http.Error(w, "self dependency not allowed", http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (g *Graphs) RemoveDependensy(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UIDInterface{}).(int64)
	graphID := r.Context().Value(GraphID{}).(int64)
	nodeID := r.Context().Value(NodeID{}).(int64)
	dependencyNodeID := r.Context().Value(DependencyNodeID{}).(int64)

	dependency := &models.Dependency{
		FromNodeID: nodeID,
		ToNodeID:   dependencyNodeID,
	}

	err := g.service.RemoveDependensy(r.Context(), userID, graphID, dependency)
	if err != nil {
		g.errorLog.Println(err)
		if errors.Is(err, service.ErrNotExists) {
			http.Error(w, "dependency not found", http.StatusBadRequest)
		}
		if errors.Is(err, service.ErrDenied) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
