package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
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

func TestCustomClaims(t *testing.T) {
	iss, aud, close := newTestIssuerAndAudience(t)
	defer close()

	newClaims := func(audience string, ttl time.Duration) *jwt.RegisteredClaims {
		return &jwt.RegisteredClaims{
			Subject:   "alice",
			Audience:  jwt.ClaimStrings{audience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
		}
	}

	t.Run("Simple", func(t *testing.T) {
		token, err := iss.SignClaims(newClaims("custom", time.Hour))
		require.NoError(t, err)

		claims := &jwt.RegisteredClaims{}
		require.NoError(t, iss.ParseClaims(token, claims))
		require.Equal(t, "alice", claims.Subject)
		require.True(t, claims.VerifyAudience("custom", true))

		// The runtime doesn't accept tokens for other audiences
		_, err = aud.ParseAndValidate(token)
		require.Error(t, err)
	})

	t.Run("Expired", func(t *testing.T) {
		token, err := iss.SignClaims(newClaims("custom", -time.Minute))
		require.NoError(t, err)
		require.Error(t, iss.ParseClaims(token, &jwt.RegisteredClaims{}))
	})

	t.Run("Other issuer", func(t *testing.T) {
		other, err := NewEphemeralIssuer("")
		require.NoError(t, err)
		token, err := other.SignClaims(newClaims("custom", time.Hour))
		require.NoError(t, err)
		require.Error(t, iss.ParseClaims(token, &jwt.RegisteredClaims{}))
	})

	t.Run("Unsigned", func(t *testing.T) {
		token, err := NewDevToken(nil, nil)
		require.NoError(t, err)
		require.Error(t, iss.ParseClaims(token, &jwt.RegisteredClaims{}))
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
