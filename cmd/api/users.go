package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/jersonsatoru/lets-go-further/internal/data"
	"github.com/jersonsatoru/lets-go-further/internal/validator"
	"go.uber.org/zap"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	v := validator.New()
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email address already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	go func(app *application) {
		err := app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			zap.S().Errorw(err.Error(), "Recipient", user.Email)
		}
	}(app)

	env := envelope{"user": user}
	b, err := json.Marshal(env)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Location", fmt.Sprintf("/v1/users/%d", user.ID))
	w.Write(b)
}
