package api

import (
	"net/http"

	"github.com/Frol333/14Sprint/pkg/middleware"
	"github.com/Frol333/14Sprint/pkg/service"

	"github.com/gorilla/mux"
)

const webDir = "./web"

func New(httpAddr string) *http.Server {
	r := mux.NewRouter()

	return &http.Server{
		Addr:    httpAddr,
		Handler: setupRouter(r),
	}
}

func setupRouter(r *mux.Router) http.Handler {
	r.StrictSlash(true)

	fs := http.FileServer(http.Dir(webDir))
	r.Handle("/", fs)
	r.HandleFunc("/api/signin", service.CheckPassword).Methods(http.MethodPost)
	api := r.PathPrefix("/api").Subrouter()
	api.Use(middleware.LoggingMiddleware)
	api.Use(middleware.Auth)
	api.HandleFunc("/tasks", service.TasksHandler).Methods(http.MethodGet)
	api.HandleFunc("/task", service.GetTask).Methods(http.MethodGet)
	api.HandleFunc("/task", service.UpdateTask).Methods(http.MethodPatch, http.MethodPut)
	api.HandleFunc("/task", service.CrateTask).Methods(http.MethodPost)
	api.HandleFunc("/task", service.DeleteTask).Methods(http.MethodDelete)
	api.HandleFunc("/task/done", service.DoneHandler).Methods(http.MethodPost)
	api.HandleFunc("/nextdate", service.NextDateHandler).Methods(http.MethodGet)

	r.PathPrefix("/").Handler(http.StripPrefix("/", fs))

	return r
}
