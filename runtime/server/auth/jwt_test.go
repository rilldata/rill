package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTokens(t *testing.T) {
	iss, aud, close := newTestIssuerAndAudience(t)
	defer close()

	t.Run("Simple", func(t *testing.T) {
		token, err := iss.NewToken(TokenOptions{
			AudienceURL:         aud.audienceURL,
			Subject:             "alice",
			TTL:                 time.Duration(time.Hour),
			SystemPermissions:   []runtime.Permission{runtime.ReadInstance},
			InstancePermissions: map[string][]runtime.Permission{"example": {runtime.ReadOLAP}},
		})

		cp, err := aud.ParseAndValidate(token)
		require.NoError(t, err)

		claims := cp.Claims("")
		require.Equal(t, "alice", claims.UserID)
		require.True(t, claims.Can(runtime.ReadInstance))
		require.False(t, claims.Can(runtime.ReadOLAP))

		claims = cp.Claims("example")
		require.True(t, claims.Can(runtime.ReadOLAP))
		require.False(t, claims.Can(runtime.ReadObjects))

		claims = cp.Claims("unknown")
		require.False(t, claims.Can(runtime.ReadOLAP))
	})

	t.Run("Expired", func(t *testing.T) {
		token, err := iss.NewToken(TokenOptions{
			AudienceURL:         aud.audienceURL,
			Subject:             "alice",
			TTL:                 time.Duration(time.Millisecond),
			SystemPermissions:   []runtime.Permission{runtime.ReadInstance},
			InstancePermissions: map[string][]runtime.Permission{"example": {runtime.ReadOLAP}},
		})

		time.Sleep(50 * time.Millisecond)

		_, err = aud.ParseAndValidate(token)
		require.Error(t, err)
	})

	t.Run("Invalid audience", func(t *testing.T) {
		token, err := iss.NewToken(TokenOptions{
			AudienceURL: "http://bad.org",
			Subject:     "alice",
			TTL:         time.Duration(time.Hour),
		})

		_, err = aud.ParseAndValidate(token)
		require.Error(t, err)
	})
}

func newTestIssuerAndAudience(t *testing.T) (*Issuer, *Audience, func()) {
	// Create Issuer and serve on a test server
	iss, err := NewEphemeralIssuer("")
	require.NoError(t, err)

	mux := http.NewServeMux()
	mux.Handle("/.well-known/jwks.json", iss.WellKnownHandler())

	srv := httptest.NewServer(mux)
	iss.issuerURL = srv.URL

	// Create Audience
	audienceURL := "http://example.org"
	aud, err := OpenAudience(context.Background(), zap.NewNop(), srv.URL, audienceURL)
	require.NoError(t, err)

	return iss, aud, func() { srv.Close(); aud.Close() }
}
