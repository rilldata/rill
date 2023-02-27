package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	iss, aud, close := newTestIssuerAndAudience(t)
	defer close()

	t.Run("Anon", func(t *testing.T) {
		handler := HTTPMiddleware(aud, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			claims := GetClaims(r.Context())
			require.NotNil(t, claims)
			require.Equal(t, "", claims.Subject())
			require.False(t, claims.Can(ManageInstances))
		})

		req := httptest.NewRequest("GET", "/", nil)

		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handler(w, r, nil) })
		httpHandler.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("Authenticated", func(t *testing.T) {
		token, err := iss.NewToken(TokenOptions{
			AudienceURL:       aud.audienceURL,
			Subject:           "token",
			TTL:               time.Hour,
			SystemPermissions: []Permission{ManageInstances},
		})
		require.NoError(t, err)

		handler := HTTPMiddleware(aud, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			claims := GetClaims(r.Context())
			require.NotNil(t, claims)
			require.Equal(t, "token", claims.Subject())
			require.True(t, claims.Can(ManageInstances))
			require.False(t, claims.Can(ReadOLAP))
		})

		req := httptest.NewRequest("GET", "/", nil)
		req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", token)}

		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handler(w, r, nil) })
		httpHandler.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("Open", func(t *testing.T) {
		// NOTE: aud is nil, indicating no authentication
		handler := HTTPMiddleware(nil, func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			claims := GetClaims(r.Context())
			require.NotNil(t, claims)
			require.Equal(t, "", claims.Subject())
			require.True(t, claims.Can(ManageInstances))
			require.True(t, claims.Can(ReadOLAP))
		})

		req := httptest.NewRequest("GET", "/", nil)

		httpHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handler(w, r, nil) })
		httpHandler.ServeHTTP(httptest.NewRecorder(), req)
	})

}
