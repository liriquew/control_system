package graphs

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type GraphID struct{}
type NodeID struct{}
type DependencyNodeID struct{}

func GraphIDGetter(next http.Handler) http.Handler {
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

func NodeIDGetter(next http.Handler) http.Handler {
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

func DependencyNodeIDGetter(next http.Handler) http.Handler {
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
