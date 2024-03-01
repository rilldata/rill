package middleware

import (
	"net/http"

	"github.com/rilldata/rill/runtime/pkg/httputil"
)

type CheckFunc func(req *http.Request) error

// Check is a middleware that only calls the next handler if the checkFn succeeds.
// If the checkFn fails, the error is written to the response as per the behavior of httputil.Handler.
func Check(checkFn CheckFunc, next http.Handler) httputil.Handler {
	return func(w http.ResponseWriter, r *http.Request) error {
		if err := checkFn(r); err != nil {
			return err
		}

		next.ServeHTTP(w, r)
		return nil
	}
}
