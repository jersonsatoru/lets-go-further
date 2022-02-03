package main

import "github.com/gorilla/mux"

func (app *application) routes() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/v1/healthcheck", app.healthCheckHandler)
	return r
}
