package appapi

import (
	"net/http"

	"github.com/liriquew/control_system/internal/api_handlers/auth"
	"github.com/liriquew/control_system/internal/api_handlers/graphs"
	"github.com/liriquew/control_system/internal/api_handlers/groups"
	"github.com/liriquew/control_system/internal/api_handlers/task"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func New(authAPI auth.AuthAPI, taskAPI task.TaskAPI, groupsAPI groups.GroupsAPI, graphsAPI graphs.GpraphsAPI) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Route("/api", func(r chi.Router) {
		r.Handle("/ping", http.HandlerFunc(Ping))

		r.Post("/signin", http.HandlerFunc(authAPI.SignIn))
		r.Post("/signup", http.HandlerFunc(authAPI.SignUp))

		r.With(authAPI.AuthRequired).Route("/tasks", func(r chi.Router) {
			r.Post("/", http.HandlerFunc(taskAPI.CreateTask))
			r.Get("/", http.HandlerFunc(taskAPI.GetTaskList))
			r.Get("/{id}", http.HandlerFunc(taskAPI.GetTask))
			r.Patch("/{id}", http.HandlerFunc(taskAPI.UpdateTask))
			r.Delete("/{id}", http.HandlerFunc(taskAPI.DeleteTask))
			r.Get("/predict/*", http.HandlerFunc(taskAPI.Predict))
		})
		r.With(authAPI.AuthRequired).Route("/groups", func(r chi.Router) {
			r.Post("/", http.HandlerFunc(groupsAPI.CreateGroup))
			r.Get("/", http.HandlerFunc(groupsAPI.ListUserGroups))

			r.With(groups.ExtractGroupID).Route("/{groupID}", func(r chi.Router) {
				r.Get("/", http.HandlerFunc(groupsAPI.GetGroup))
				r.Delete("/", http.HandlerFunc(groupsAPI.DeleteGroup))
				r.Patch("/", http.HandlerFunc(groupsAPI.UpdateGroup))

				r.Route("/members", func(r chi.Router) {
					r.Post("/", http.HandlerFunc(groupsAPI.AddGroupMember))
					r.Get("/", http.HandlerFunc(groupsAPI.ListGroupMembers))
					r.With(groups.ExtractMemberID).Delete("/{memberID}", http.HandlerFunc(groupsAPI.RemoveGroupMember))
					r.With(groups.ExtractMemberID).Patch("/{memberID}/role", http.HandlerFunc(groupsAPI.ChangeMemberRole))
				})
				r.Route("/graphs", func(r chi.Router) {
					r.Post("/", http.HandlerFunc(groupsAPI.CreateGroupGraph))
					r.Get("/", http.HandlerFunc(groupsAPI.ListGroupGraphs))
				})
			})
		})

		r.With(authAPI.AuthRequired, graphs.GraphIDGetter).Route("/graphs/{graphID}", func(r chi.Router) {
			r.Get("/", http.HandlerFunc(graphsAPI.GetGraph))
			r.Get("/predict", http.HandlerFunc(graphsAPI.PredictGraph))

			r.Route("/nodes", func(r chi.Router) {
				r.Post("/", http.HandlerFunc(graphsAPI.CreateNode))

				r.With(graphs.NodeIDGetter).Route("/{nodeID}", func(r chi.Router) {
					r.Get("/", http.HandlerFunc(graphsAPI.GetNode))
					r.Delete("/", http.HandlerFunc(graphsAPI.RemoveNode))
					r.Patch("/", http.HandlerFunc(graphsAPI.UpdateNode))

					r.Route("/dependencies", func(r chi.Router) {
						r.Get("/", http.HandlerFunc(graphsAPI.GetDependencies))
						r.With(graphs.DependencyNodeIDGetter).Route("/{dependencyNodeID}", func(r chi.Router) {
							r.Post("/", http.HandlerFunc(graphsAPI.AddDependency))
							r.Delete("/", http.HandlerFunc(graphsAPI.RemoveDependensy))
						})
					})
				})

			})
		})

	})

	return r
}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
