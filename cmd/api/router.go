package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() *mux.Router {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(app.notFoundErrorResponse)
	r.MethodNotAllowedHandler = http.HandlerFunc(app.methodNotAllowedError)
	r.Handle("/v1/healthcheck", app.rateLimit(http.HandlerFunc(app.healthCheckHandler)))
	r.Handle("/v1/movies", app.rateLimit(http.HandlerFunc(app.listMoviesHandler))).Methods(http.MethodGet)
	r.Handle("/v1/movies", app.rateLimit(http.HandlerFunc(app.createMovieHandler))).Methods(http.MethodPost)
	r.Handle("/v1/movies/{id:[0-9]+}", app.rateLimit(http.HandlerFunc(app.updateMovieHandler))).Methods(http.MethodPut)
	r.Handle("/v1/movies/{id:[0-9]+}", app.rateLimit(http.HandlerFunc(app.partialUpdateMovieHandler))).Methods(http.MethodPatch)
	r.Handle("/v1/movies/{id:[0-9]+}", app.rateLimit(http.HandlerFunc(app.showMovieHandler))).Methods(http.MethodGet)
	r.Handle("/v1/movies/{id:[0-9]+}", app.rateLimit(http.HandlerFunc(app.deleteMovieHandler))).Methods(http.MethodDelete)
	r.Handle("/v1/users", app.rateLimit(http.HandlerFunc(app.registerUserHandler))).Methods(http.MethodPost)
	r.Handle("/v1/users/activated", app.rateLimit(http.HandlerFunc(app.activateUserHandler))).Methods(http.MethodPut)
	r.Handle("/v1/tokens/authentication", app.rateLimit(http.HandlerFunc(app.createAuthenticationTokenHandler))).Methods(http.MethodPost)

	r.Use(app.recoverPanic)
	r.Use(app.rateLimit)
	return r
}
