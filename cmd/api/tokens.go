package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jersonsatoru/lets-go-further/internal/data"
	"github.com/jersonsatoru/lets-go-further/internal/validator"
)

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	data.ValidateEmail(v, input.Email)
	data.ValidatePasswordPlaintext(v, input.Password)
	if !v.Valid() {
		app.invalidCredentialsResponse(w, r)
		return
	}

	user, err := app.models.Users.GetByEmail(input.Email)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidCredentialsResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	b, err := user.Password.Matches(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !b {
		app.invalidCredentialsResponse(w, r)
		return
	}
	err = app.models.Tokens.DeleteAllForUser(data.ScopedAuthentication, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	token, err := app.models.Tokens.New(user.ID, time.Duration(time.Hour*3*24), data.ScopedAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	content, err := json.Marshal(token)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(content))
}
