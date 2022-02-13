package main

import (
	"expvar"
	"net/http"

	"github.com/gorilla/mux"
)

func (app *application) routes() *mux.Router {
	r := mux.NewRouter()
	r.NotFoundHandler = http.HandlerFunc(app.notFoundErrorResponse)
	r.MethodNotAllowedHandler = http.HandlerFunc(app.methodNotAllowedError)
	r.Handle("/v1/healthcheck", app.rateLimit(http.HandlerFunc(app.healthCheckHandler)))

	r.Handle("/v1/movies", app.requirePermission(app.rateLimit(http.HandlerFunc(app.listMoviesHandler)), "movies:read")).Methods(http.MethodGet, http.MethodOptions)
	r.Handle("/v1/movies", app.requirePermission(app.rateLimit(http.HandlerFunc(app.createMovieHandler)), "movies:write")).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/v1/movies/{id:[0-9]+}", app.requirePermission(app.rateLimit(http.HandlerFunc(app.updateMovieHandler)), "movies:write")).Methods(http.MethodPut, http.MethodOptions)
	r.Handle("/v1/movies/{id:[0-9]+}", app.requirePermission(app.rateLimit(http.HandlerFunc(app.partialUpdateMovieHandler)), "movies:write")).Methods(http.MethodPatch, http.MethodOptions)
	r.Handle("/v1/movies/{id:[0-9]+}", app.requirePermission(app.rateLimit(http.HandlerFunc(app.showMovieHandler)), "movies:read")).Methods(http.MethodGet, http.MethodOptions)
	r.Handle("/v1/movies/{id:[0-9]+}", app.requirePermission(app.rateLimit(http.HandlerFunc(app.deleteMovieHandler)), "movies:write")).Methods(http.MethodDelete, http.MethodOptions)

	r.Handle("/v1/users", app.rateLimit(http.HandlerFunc(app.registerUserHandler))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/v1/users/activated", app.rateLimit(http.HandlerFunc(app.activateUserHandler))).Methods(http.MethodPut, http.MethodOptions)
	r.Handle("/v1/tokens/authentication", app.metrics(app.rateLimit(http.HandlerFunc(app.createAuthenticationTokenHandler)))).Methods(http.MethodPost, http.MethodOptions)
	r.Handle("/v1/metrics", app.metrics(expvar.Handler()))
	r.Use(app.recoverPanic)
	r.Use(app.authenticate)
	r.Use(app.enableCORS)
	return r
}
