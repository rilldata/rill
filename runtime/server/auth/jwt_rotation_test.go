package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAudienceRefreshesRotatedJWKS(t *testing.T) {
	// An unknown key ID must trigger one synchronous refresh so a normal key
	// rotation does not require a process restart or a timing-based retry loop.
	first, err := NewEphemeralIssuer("")
	require.NoError(t, err)
	second, err := NewEphemeralIssuer("")
	require.NoError(t, err)

	fixture := newMutableJWKSServer(t, first.publicJWKS)
	first.issuerURL = fixture.server.URL
	second.issuerURL = fixture.server.URL
	audienceURL := "https://runtime.example.com"
	audience, err := OpenAudience(t.Context(), zap.NewNop(), fixture.server.URL, audienceURL)
	require.NoError(t, err)
	t.Cleanup(audience.Close)

	oldToken, err := first.NewToken(TokenOptions{AudienceURL: audienceURL, Subject: "alice", TTL: time.Hour})
	require.NoError(t, err)
	_, err = audience.ParseAndValidate(oldToken)
	require.NoError(t, err)

	fixture.set(http.StatusOK, second.publicJWKS)
	rotatedToken, err := second.NewToken(TokenOptions{AudienceURL: audienceURL, Subject: "alice", TTL: time.Hour})
	require.NoError(t, err)
	_, err = audience.ParseAndValidate(rotatedToken)
	require.NoError(t, err)
	require.Equal(t, 2, fixture.requests())
	require.Equal(t, []string{second.signingKey.KeyID}, audience.jwks.KIDs())

	// Once the issuer has removed the old key, tokens signed by it must fail closed.
	_, err = audience.ParseAndValidate(oldToken)
	require.Error(t, err)
}

func TestAudienceRejectsMalformedJWKSRefreshAndKeepsLastGoodKeys(t *testing.T) {
	// A malformed refresh is untrusted network input. It must not admit the
	// unknown token or erase the last known-good key used by existing tokens.
	trusted, err := NewEphemeralIssuer("")
	require.NoError(t, err)
	untrusted, err := NewEphemeralIssuer("")
	require.NoError(t, err)
	fixture := newMutableJWKSServer(t, trusted.publicJWKS)
	trusted.issuerURL = fixture.server.URL
	untrusted.issuerURL = fixture.server.URL
	audienceURL := "https://runtime.example.com"
	audience, err := OpenAudience(t.Context(), zap.NewNop(), fixture.server.URL, audienceURL)
	require.NoError(t, err)
	t.Cleanup(audience.Close)

	trustedToken, err := trusted.NewToken(TokenOptions{AudienceURL: audienceURL, Subject: "alice", TTL: time.Hour})
	require.NoError(t, err)
	untrustedToken, err := untrusted.NewToken(TokenOptions{AudienceURL: audienceURL, Subject: "mallory", TTL: time.Hour})
	require.NoError(t, err)

	fixture.set(http.StatusOK, []byte(`{"keys":[`))
	_, err = audience.ParseAndValidate(untrustedToken)
	require.Error(t, err)
	_, err = audience.ParseAndValidate(trustedToken)
	require.NoError(t, err)
	require.Equal(t, []string{trusted.signingKey.KeyID}, audience.jwks.KIDs())
}

func TestAudienceValidatesIssuerAndAudience(t *testing.T) {
	// Valid signatures with routing claims for another issuer or runtime must still be rejected.
	issuer, audience, close := newTestIssuerAndAudience(t)
	t.Cleanup(close)

	// Signature validity is insufficient: both routing claims bind a token to
	// the intended control plane and runtime.
	t.Run("issuer mismatch", func(t *testing.T) {
		original := issuer.issuerURL
		issuer.issuerURL = "https://different-issuer.example.com"
		t.Cleanup(func() { issuer.issuerURL = original })
		token, err := issuer.NewToken(TokenOptions{AudienceURL: audience.audienceURL, Subject: "alice", TTL: time.Hour})
		require.NoError(t, err)
		_, err = audience.ParseAndValidate(token)
		require.ErrorContains(t, err, "invalid token issuer")
	})

	t.Run("audience mismatch", func(t *testing.T) {
		token, err := issuer.NewToken(TokenOptions{AudienceURL: "https://different-runtime.example.com", Subject: "alice", TTL: time.Hour})
		require.NoError(t, err)
		_, err = audience.ParseAndValidate(token)
		require.Error(t, err)
	})
}

func TestOpenAudienceInitialFetchFailuresHonorContext(t *testing.T) {
	// Startup retry must stop on cancellation instead of imposing the full
	// twenty-attempt backoff when the caller is already shutting down.
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{name: "server error", status: http.StatusInternalServerError, body: "unavailable"},
		{name: "malformed JWKS", status: http.StatusOK, body: `{"keys":[`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
				cancel()
			}))
			t.Cleanup(server.Close)

			started := time.Now()
			audience, err := OpenAudience(ctx, zap.NewNop(), server.URL, "https://runtime.example.com")
			require.Error(t, err)
			require.Nil(t, audience)
			require.Less(t, time.Since(started), time.Second)
		})
	}
}

type mutableJWKSServer struct {
	server *httptest.Server

	mu      sync.RWMutex
	status  int
	body    []byte
	request int
}

func newMutableJWKSServer(t *testing.T, body []byte) *mutableJWKSServer {
	t.Helper()
	fixture := &mutableJWKSServer{status: http.StatusOK, body: append([]byte(nil), body...)}
	fixture.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fixture.mu.Lock()
		fixture.request++
		status := fixture.status
		response := append([]byte(nil), fixture.body...)
		fixture.mu.Unlock()
		w.WriteHeader(status)
		_, _ = w.Write(response)
	}))
	t.Cleanup(fixture.server.Close)
	return fixture
}

func (f *mutableJWKSServer) set(status int, body []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.status = status
	f.body = append([]byte(nil), body...)
}

func (f *mutableJWKSServer) requests() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.request
}
