package graphs

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type GraphID struct{}
type NodeID struct{}
type DependencyNodeID struct{}

type GraphsMiddleware struct {
	log *slog.Logger
}

func NewGraphsMiddleware(log *slog.Logger) *GraphsMiddleware {
	return &GraphsMiddleware{
		log: log,
	}
}

func (g *GraphsMiddleware) GraphIDGetter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		graphID, err := strconv.ParseInt(chi.URLParam(r, "graphID"), 10, 64)
		if err != nil {
			http.Error(w, "graphID path param required", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), GraphID{}, graphID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *GraphsMiddleware) NodeIDGetter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nodeID, err := strconv.ParseInt(chi.URLParam(r, "nodeID"), 10, 64)
		if err != nil {
			http.Error(w, "nodeID path param required", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), NodeID{}, nodeID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (g *GraphsMiddleware) DependencyNodeIDGetter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nodeID, err := strconv.ParseInt(chi.URLParam(r, "dependencyNodeID"), 10, 64)
		if err != nil {
			http.Error(w, "nodeID path param required", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), DependencyNodeID{}, nodeID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetGraphID(r *http.Request) int64 {
	id, _ := r.Context().Value(GraphID{}).(int64)
	return id
}

func GetNodeID(r *http.Request) int64 {
	id, _ := r.Context().Value(NodeID{}).(int64)
	return id
}

func GetDependencyNodeID(r *http.Request) int64 {
	id, _ := r.Context().Value(DependencyNodeID{}).(int64)
	return id
}

func GetPadding(r *http.Request) int64 {
	padding, _ := strconv.ParseInt(r.URL.Query().Get("padding"), 10, 64)
	return padding
}
