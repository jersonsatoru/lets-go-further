package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jersonsatoru/lets-go-further/internal/data"
	"github.com/jersonsatoru/lets-go-further/internal/validator"
	"go.uber.org/zap"
)

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		app.notFoundErrorResponse(w, r)
		return
	}
	if id <= 0 {
		zap.L().Error("error to decode json data", zap.Int("id", id))
		app.notFoundErrorResponse(w, r)
		return
	}
	data := &data.Movie{
		ID:        1,
		CreatedAt: time.Now(),
		Title:     "Independence day",
		Year:      2002,
		Runtime:   120,
		Genres:    []string{"Terror", "Adventure"},
		Version:   1,
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": data}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	movie := &data.Movie{
		ID:        1,
		Title:     input.Title,
		Year:      input.Year,
		Runtime:   data.Runtime(input.Runtime),
		Genres:    input.Genres,
		Version:   1,
		CreatedAt: time.Now(),
	}

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	env := envelope{
		"movie": movie,
	}
	var header http.Header = http.Header{}
	header.Add("location", fmt.Sprintf("%s/movies/1", r.Host))
	app.writeJSON(w, http.StatusCreated, env, header)
}
