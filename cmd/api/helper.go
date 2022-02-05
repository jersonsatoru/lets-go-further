package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	for key, value := range headers {
		w.Header()[key] = value
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(b))
	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarchalTypeError *json.UnmarshalTypeError
		var invalidUnmarchalError *json.InvalidUnmarshalError
		switch {
		case errors.As(err, &syntaxError):
			message := fmt.Sprintf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
			return fmt.Errorf(message)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("body contains badly-formed JSON")
		case errors.As(err, &unmarchalTypeError):
			if unmarchalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrectly JSON type for field %q", unmarchalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrectly JSON type (at character %d)", unmarchalTypeError.Offset)
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarchalError):
			panic(err)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown field %s", field)
		case err.Error() == "http: request body too large":
			return fmt.Errorf("bodu must not be larger than %d bytes", maxBytes)
		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}
	return nil
}
