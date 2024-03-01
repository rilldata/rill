package httputil

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Handler is similar to http.HandlerFunc with simple error handling built in.
// If an error is returned, it will be written to the response as a JSON object (using WriteError).
type Handler func(http.ResponseWriter, *http.Request) error

func (fn Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := fn(w, r)
	if err != nil {
		WriteError(w, err)
	}
}

// Error creates a new HTTP error with a status code from an existing error.
func Error(statusCode int, err error) error {
	return HTTPError{
		StatusCode: statusCode,
		Err:        err,
	}
}

// Errorf creates a new HTTP error with a status code and formatted message message.
func Errorf(statusCode int, format string, args ...interface{}) error {
	return HTTPError{
		StatusCode: statusCode,
		Err:        fmt.Errorf(format, args...),
	}
}

// HTTPError represents an error with a HTTP status code.
type HTTPError struct {
	StatusCode int
	Err        error
}

func (e HTTPError) Error() string {
	return e.Err.Error()
}

func (e HTTPError) Unwrap() error {
	return e.Err
}

// WriteError writes an error to the response as a JSON object.
// If the error is an instance of Error, the status code will be set to the error's code.
// Otherwise, it will return a 400 status code.
func WriteError(w http.ResponseWriter, err error) {
	// Set content type to JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Write the error code
	code := http.StatusBadRequest
	var httperr HTTPError
	if errors.As(err, &httperr) {
		code = httperr.StatusCode
	}
	w.WriteHeader(code)

	// Write error as JSON
	obj := map[string]string{"error": err.Error()}
	json, err := json.Marshal(obj)
	if err != nil {
		panic(err)
	}
	w.Write(json)
}
