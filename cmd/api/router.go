package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() *mux.Router {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(app.notFoundErrorResponse)
	r.MethodNotAllowedHandler = http.HandlerFunc(app.methodNotAllowedError)
	r.HandleFunc("/v1/healthcheck", app.healthCheckHandler)
	r.HandleFunc("/v1/movies/{id:[0-9]+}", app.showMovieHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/movies", app.createMovieHandler).Methods(http.MethodPost)
	return r
}
