package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

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

	token, err := app.models.Tokens.New(user.ID, time.Duration(time.Hour*3*24), data.ScopedActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.wg.Add(1)
	go func(app *application) {
		defer app.wg.Done()
		defer func() {
			if err := recover(); err != nil {
				zap.S().Errorw(fmt.Sprintf("%v", err), "Recipient", user.Email)
			}
		}()
		time.Sleep(3 * time.Second)
		data := map[string]interface{}{
			"userID":          user.ID,
			"activationToken": token.Plaintext,
		}
		err := app.mailer.Send(user.Email, "user_welcome.tmpl", data)
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

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TokenPlaintext string `json:"token"`
	}
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	v := validator.New()
	if data.ValidateToken(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.models.Users.GetForToken(input.TokenPlaintext, data.ScopedActivation)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w, r, v.Errors)
			return
		}
	}
	user.Activated = true
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	err = app.models.Tokens.DeleteAllForUser(data.ScopedActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = json.NewEncoder(w).Encode(user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
