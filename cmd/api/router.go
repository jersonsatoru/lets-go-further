package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/v1/healthcheck", app.healthCheckHandler)
	r.HandleFunc("/v1/movie/{id:[0-9]+}", app.showMovieHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/movie/", app.createMovieHandler)
	return r
}
