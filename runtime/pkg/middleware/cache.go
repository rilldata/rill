package middleware

import "net/http"

// CacheControlMiddleware sets a default Cache-Control header of "no-store, no-cache, must-revalidate".
// Handlers that need different caching behavior (e.g. static assets) can override it by calling
// w.Header().Set("Cache-Control", ...) before their first write.
func CacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		next.ServeHTTP(w, r)
	})
}
