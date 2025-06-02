package task

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	tasksclient "github.com/liriquew/control_system/api/internal/grpc/clients/tasks"
)

type TaskID struct{}

type TasksMiddleware struct {
	log    *slog.Logger
	client tasksclient.TasksClient
}

func NewTasksMiddleware(log *slog.Logger, client tasksclient.TasksClient) *TasksMiddleware {
	return &TasksMiddleware{
		log:    log,
		client: client,
	}
}

func (g *TasksMiddleware) ExtractTaskID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		taskID, err := strconv.ParseInt(chi.URLParam(r, "taskID"), 10, 64)
		g.log.Debug("taskID in middleware ExtractTaskID", slog.Int64("taskID", taskID))
		if err != nil {
			http.Error(w, "taskID path param required", http.StatusBadRequest)
			return
		}
		if taskID <= 0 {
			http.Error(w, "invalid taskID", http.StatusBadRequest)
			return
		}

		ctx := context.WithValue(r.Context(), TaskID{}, taskID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetTaskID(r *http.Request) int64 {
	id, _ := r.Context().Value(TaskID{}).(int64)
	return id
}

func GetOffset(r *http.Request) int64 {
	offset, _ := strconv.ParseInt(r.URL.Query().Get("offset"), 10, 64)
	return offset
}
