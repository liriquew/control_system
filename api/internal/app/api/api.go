package appapi

import (
	"net/http"
	"time_manage/internal/api_handlers/auth"
	"time_manage/internal/api_handlers/task"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

func New(authAPI auth.AuthAPI, taskAPI task.TaskAPI) *chi.Mux {
	r := chi.NewRouter()

	// если где-то что-то пойдет не так (произойдет паника) то все не упадет
	r.Use(middleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		r.Handle("/ping", http.HandlerFunc(Ping))

		r.Post("/signin", http.HandlerFunc(authAPI.SignIn))
		r.Post("/signup", http.HandlerFunc(authAPI.SignUp))

		r.With(authAPI.AuthRequired).Route("/task", func(r chi.Router) {
			r.Post("/*", http.HandlerFunc(taskAPI.CreateTask))
			r.Get("/*", http.HandlerFunc(taskAPI.GetTask))
			r.Patch("/*", http.HandlerFunc(taskAPI.UpdateTask))
			r.Delete("/*", http.HandlerFunc(taskAPI.DeleteTask))
		})

		r.With(authAPI.AuthRequired).Get("/predict/*", http.HandlerFunc(taskAPI.Predict))
	})

	return r
}

func Ping(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Pong"))
}
