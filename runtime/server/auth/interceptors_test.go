package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
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

	t.Run("OpenBehindAuthProxy", func(t *testing.T) {
		// An auth proxy in front of Rill Developer adds its own JWT, which has no Rill permissions.
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"sub": "proxy-user", "email": "user@example.com"}).SignedString([]byte("proxy-secret"))
		require.NoError(t, err)

		handler := HTTPMiddleware(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context(), "")
			require.NotNil(t, claims)
			require.True(t, claims.Admin())
			require.True(t, claims.SkipChecks)
			require.True(t, claims.Can(runtime.ReadObjects))
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", token)}

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("OpenWithDevJWT", func(t *testing.T) {
		// A dev JWT from IssueDevJWT ("View as") applies the mocked user's attributes to security policies.
		token, err := NewDevToken(map[string]any{"admin": false, "email": "viewer@example.com"}, runtime.AllPermissions)
		require.NoError(t, err)

		handler := HTTPMiddleware(nil, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaims(r.Context(), "")
			require.NotNil(t, claims)
			require.False(t, claims.Admin())
			require.False(t, claims.SkipChecks)
			require.Equal(t, "viewer@example.com", claims.UserAttributes["email"])
			require.True(t, claims.Can(runtime.ReadObjects))
		}))

		req := httptest.NewRequest("GET", "/", nil)
		req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", token)}

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusOK, recorder.Code)
	})
}
