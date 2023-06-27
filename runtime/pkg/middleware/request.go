package middleware

import (
	"errors"
	"net/http"
)

// RequestHTTPHandler calls fn on each request.
func RequestHTTPHandler(route string, fn func(route string, req *http.Request) error, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(route, r); err != nil {
			var httpError *HTTPError
			if errors.As(err, &httpError) {
				http.Error(w, err.Error(), httpError.Code)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

type HTTPError struct {
	Code    int
	Message string
}

func NewHTTPError(code int, msg string) *HTTPError {
	return &HTTPError{code, msg}
}

func (e *HTTPError) Error() string {
	return e.Message
}
