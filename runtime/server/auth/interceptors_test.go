package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/stretchr/testify/require"
)

func TestMiddleware(t *testing.T) {
	iss, aud, close := newTestIssuerAndAudience(t)
	defer close()

	t.Run("Anon", func(t *testing.T) {
		handler := HTTPMiddleware(aud, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context(), "")
			require.NotNil(t, claims)
			require.Equal(t, "", claims.UserID)
			require.False(t, claims.Can(runtime.ManageInstances))
		}))

		req := httptest.NewRequest("GET", "/", nil)

		handler.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("Authenticated", func(t *testing.T) {
		token, err := iss.NewToken(TokenOptions{
			AudienceURL:       aud.audienceURL,
			Subject:           "token",
			TTL:               time.Hour,
			SystemPermissions: []runtime.Permission{runtime.ReadInstance},
		})
		require.NoError(t, err)

		handler := HTTPMiddleware(aud, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context(), "")
			require.NotNil(t, claims)
			require.Equal(t, "token", claims.UserID)
			require.True(t, claims.Can(runtime.ReadInstance))
			require.False(t, claims.Can(runtime.ReadOLAP))
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", token)}

		handler.ServeHTTP(httptest.NewRecorder(), req)
	})

	t.Run("Open", func(t *testing.T) {
		// NOTE: aud is nil, indicating no authentication
		handler := HTTPMiddleware(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context(), "")
			require.NotNil(t, claims)
			require.Equal(t, "", claims.UserID)
			require.True(t, claims.Can(runtime.ManageInstances))
			require.True(t, claims.Can(runtime.ReadOLAP))
		}))

		req := httptest.NewRequest("GET", "/", nil)

		handler.ServeHTTP(httptest.NewRecorder(), req)
	})

}
