package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jersonsatoru/lets-go-further/internal/data"
	"go.uber.org/zap"
)

func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		zap.L().Error("error to decode json data", zap.Int("id", id))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if id <= 0 {
		zap.L().Error("error to decode json data", zap.Int("id", id))
		http.Error(w, "invalid id", http.StatusBadRequest)
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
		zap.L().Error("error to decode json data")
		http.Error(w, "error to encode json data", http.StatusInternalServerError)
	}
}

func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("location", fmt.Sprintf("%s/movie/1", r.Host))
	w.WriteHeader(http.StatusCreated)
}
