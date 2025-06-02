package appapi

import (
	"net/http"

	"github.com/liriquew/control_system/api/internal/api_handlers/auth"
	"github.com/liriquew/control_system/api/internal/api_handlers/graphs"
	"github.com/liriquew/control_system/api/internal/api_handlers/groups"
	"github.com/liriquew/control_system/api/internal/api_handlers/task"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func New(
	authAPI auth.AuthService,
	taskAPI task.TasksService,
	groupsAPI groups.GroupsService,
	graphsAPI graphs.GraphsService,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.Logger)

	r.Route("/api", func(r chi.Router) {
		r.Handle("/ping", http.HandlerFunc(Ping))

		r.Post("/signin", http.HandlerFunc(authAPI.SignIn))
		r.Post("/signup", http.HandlerFunc(authAPI.SignUp))

		r.With(authAPI.Authenticate).Route("/tasks", func(r chi.Router) {
			r.Route("/tags", func(r chi.Router) {
				r.Get("/", http.HandlerFunc(taskAPI.GetTags))
				r.Get("/predict", http.HandlerFunc(taskAPI.PredictTags))
			})
			r.Post("/", http.HandlerFunc(taskAPI.CreateTask))
			r.Get("/", http.HandlerFunc(taskAPI.GetTaskList))

			r.With(taskAPI.ExtractTaskID).Route("/{taskID}", func(r chi.Router) {
				r.Get("/", http.HandlerFunc(taskAPI.GetTask))
				r.Patch("/", http.HandlerFunc(taskAPI.UpdateTask))
				r.Delete("/", http.HandlerFunc(taskAPI.DeleteTask))
				r.Get("/predict", http.HandlerFunc(taskAPI.PredictTask))
			})
			r.Get("/predict", http.HandlerFunc(taskAPI.PredictUncreatedTask))
		})

		r.With(authAPI.Authenticate).Route("/groups", func(r chi.Router) {
			r.Post("/", http.HandlerFunc(groupsAPI.CreateGroup))
			r.Get("/", http.HandlerFunc(groupsAPI.ListUserGroups))

			r.Route("/{groupID}", func(r chi.Router) {
				r.With(groupsAPI.CheckMemberPermission).Get("/", http.HandlerFunc(groupsAPI.GetGroup))
				r.With(groupsAPI.CheckEditorPermission).Patch("/", http.HandlerFunc(groupsAPI.UpdateGroup))
				r.With(groupsAPI.CheckAdminPermission).Delete("/", http.HandlerFunc(groupsAPI.DeleteGroup))

				r.Route("/tasks", func(r chi.Router) {
					r.With(groupsAPI.CheckEditorPermission).Post("/", http.HandlerFunc(taskAPI.CreateTask))

					r.Route("/{taskID}", func(r chi.Router) {
						r.Use(taskAPI.ExtractTaskID)
						r.With(groupsAPI.CheckMemberPermission).Get("/", http.HandlerFunc(taskAPI.GetTask))
						r.With(groupsAPI.CheckEditorPermission).Patch("/", http.HandlerFunc(taskAPI.UpdateTask))
						r.With(groupsAPI.CheckEditorPermission).Delete("/", http.HandlerFunc(taskAPI.DeleteTask))
						r.With(groupsAPI.CheckEditorPermission).Get("/predict", http.HandlerFunc(taskAPI.PredictTask))
					})

					r.With(
						groupsAPI.CheckEditorPermission,
						groupsAPI.CheckPredictedUserMemberPermission,
					).Get("/predict/{predictedUserID}", http.HandlerFunc(taskAPI.PredictUncreatedTask))
				})

				r.Route("/members", func(r chi.Router) {
					r.With(groupsAPI.CheckAdminPermission).Post("/", http.HandlerFunc(groupsAPI.AddGroupMember))
					r.With(groupsAPI.CheckMemberPermission).Get("/", http.HandlerFunc(groupsAPI.ListGroupMembers))
					r.With(groupsAPI.CheckAdminPermission, groupsAPI.ExtractMemberID).
						Delete("/{memberID}", http.HandlerFunc(groupsAPI.RemoveGroupMember))
					r.With(groupsAPI.CheckAdminPermission, groupsAPI.ExtractMemberID).
						Patch("/{memberID}/role", http.HandlerFunc(groupsAPI.ChangeMemberRole))
				})

				r.Route("/graphs", func(r chi.Router) {
					r.With(groupsAPI.CheckAdminPermission).Post("/", http.HandlerFunc(graphsAPI.CreateGroupGraph))
					r.With(groupsAPI.CheckMemberPermission).Get("/", http.HandlerFunc(graphsAPI.ListGroupGraphs))

					r.With(graphsAPI.GraphIDGetter).Route("/{graphID}", func(r chi.Router) {
						r.With(groupsAPI.CheckMemberPermission).Get("/", http.HandlerFunc(graphsAPI.GetGraph))
						r.With(groupsAPI.CheckMemberPermission).Get("/predict", http.HandlerFunc(graphsAPI.PredictGraph))

						r.Route("/nodes", func(r chi.Router) {
							r.With(groupsAPI.CheckEditorPermission).Post("/", http.HandlerFunc(graphsAPI.CreateNode))

							r.With(graphsAPI.NodeIDGetter).Route("/{nodeID}", func(r chi.Router) {
								r.With(groupsAPI.CheckMemberPermission).Get("/", http.HandlerFunc(graphsAPI.GetNode))
								r.With(groupsAPI.CheckEditorPermission).Delete("/", http.HandlerFunc(graphsAPI.RemoveNode))
								r.With(groupsAPI.CheckEditorPermission).Patch("/", http.HandlerFunc(graphsAPI.UpdateNode))

								r.Route("/dependencies", func(r chi.Router) {
									r.With(groupsAPI.CheckMemberPermission).Get("/", http.HandlerFunc(graphsAPI.GetDependencies))
									r.With(graphsAPI.DependencyNodeIDGetter).Route("/{dependencyNodeID}", func(r chi.Router) {
										r.With(groupsAPI.CheckEditorPermission).Post("/", http.HandlerFunc(graphsAPI.AddDependency))
										r.With(groupsAPI.CheckEditorPermission).Delete("/", http.HandlerFunc(graphsAPI.RemoveDependency))
									})
								})
							})
						})
					})
				}) // graphs
			}) // {groupID}
		}) // groups
	}) // api

	return r
}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
