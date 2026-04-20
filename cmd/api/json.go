package main

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New()
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578 //set max response to 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	decoder := json.NewDecoder((r.Body))
	decoder.DisallowUnknownFields() //don't allow unknown fields in data structure when unmarshalling

	return decoder.Decode(data)
}

// When building api, need consistence
// So creating this method and always use envelope struct
func writeJSONError(w http.ResponseWriter, status int, message string) error {
	type envelope struct {
		Error string `json:"error"`
	}

	return writeJSON(w, status, &envelope{Error: message})
}

func (app *application) jsonResponse(w http.ResponseWriter, status int, data any) error {
	type envelope struct {
		Data any `json:"data"`
	}

	return writeJSON(w, status, &envelope{Data: data})
}
