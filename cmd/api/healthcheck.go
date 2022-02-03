package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"port":       app.cfg.port,
		"env":        app.cfg.env,
		"apiVersion": version,
		"status":     "available",
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
