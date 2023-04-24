package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
			SystemPermissions:   []Permission{ManageInstances},
			InstancePermissions: map[string][]Permission{"example": {ReadOLAP}},
		})

		claims, err := aud.ParseAndValidate(token)
		require.NoError(t, err)
		require.Equal(t, "alice", claims.Subject())
		require.True(t, claims.Can(ManageInstances))
		require.False(t, claims.Can(ReadOLAP))
		require.True(t, claims.CanInstance("example", ReadOLAP))
		require.False(t, claims.CanInstance("example", ReadObjects))
		require.False(t, claims.CanInstance("unknown", ReadOLAP))
	})

	t.Run("Expired", func(t *testing.T) {
		token, err := iss.NewToken(TokenOptions{
			AudienceURL:         aud.audienceURL,
			Subject:             "alice",
			TTL:                 time.Duration(time.Millisecond),
			SystemPermissions:   []Permission{ManageInstances},
			InstancePermissions: map[string][]Permission{"example": {ReadOLAP}},
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
	aud, err := OpenAudience(zap.NewNop(), srv.URL, audienceURL)
	require.NoError(t, err)

	return iss, aud, func() { srv.Close(); aud.Close() }
}
