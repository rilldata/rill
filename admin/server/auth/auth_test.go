package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-jose/go-jose/v3"
	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/rilldata/rill/admin"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testProvider is a minimal OIDC provider serving a discovery document and a JWKS, and signing ID tokens with its key.
type testProvider struct {
	*httptest.Server
	key *rsa.PrivateKey
}

// newTestProvider starts a testProvider. The discovery document can be extended (or fields overridden) with extra.
func newTestProvider(t *testing.T, extra map[string]any) *testProvider {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	p := &testProvider{key: key}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		doc := map[string]any{
			"issuer":                                p.URL,
			"authorization_endpoint":                p.URL + "/authorize",
			"token_endpoint":                        p.URL + "/token",
			"jwks_uri":                              p.URL + "/jwks",
			"id_token_signing_alg_values_supported": []string{"RS256"},
		}
		for k, v := range extra {
			doc[k] = v
		}
		_ = json.NewEncoder(w).Encode(doc)
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{Key: &key.PublicKey, KeyID: "test", Algorithm: "RS256", Use: "sig"}}})
	})
	p.Server = httptest.NewServer(mux)
	t.Cleanup(p.Close)
	return p
}

// signIDToken returns an ID token for the given claims, signed with key (the provider's own key if nil).
func (p *testProvider) signIDToken(t *testing.T, key *rsa.PrivateKey, claims map[string]any) string {
	if key == nil {
		key = p.key
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithHeader("kid", "test"))
	require.NoError(t, err)
	raw, err := jwt.Signed(signer).Claims(claims).CompactSerialize()
	require.NoError(t, err)
	return raw
}

func TestIssuerURL(t *testing.T) {
	tests := []struct {
		authDomain string
		want       string
	}{
		{"rill.auth0.com", "https://rill.auth0.com/"},
		{"https://idp.example.com/realms/rill", "https://idp.example.com/realms/rill"},
		{"https://idp.example.com/realms/rill/", "https://idp.example.com/realms/rill/"},
		{"http://localhost:5556/dex", "http://localhost:5556/dex"},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, issuerURL(tt.authDomain), tt.authDomain)
	}
}

func TestNewAuthenticator(t *testing.T) {
	urls, err := admin.NewURLs("http://localhost:8080", "http://localhost:3000")
	require.NoError(t, err)
	adm := &admin.Service{URLs: urls}

	tests := []struct {
		name               string
		endSession         any // value of end_session_endpoint in the discovery document; nil leaves it out
		wantEndSession     string
		wantErrorSubstring string
	}{
		{name: "with end_session_endpoint", endSession: "https://idp.example.com/logout", wantEndSession: "https://idp.example.com/logout"},
		{name: "without end_session_endpoint"},
		{name: "malformed end_session_endpoint", endSession: 42, wantErrorSubstring: "discovery document"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var discovery map[string]any
			if tt.endSession != nil {
				discovery = map[string]any{"end_session_endpoint": tt.endSession}
			}
			p := newTestProvider(t, discovery)

			a, err := NewAuthenticator(zap.NewNop(), adm, nil, &AuthenticatorOptions{AuthDomain: p.URL, AuthClientID: "rill-client"})
			if tt.wantErrorSubstring != "" {
				require.ErrorContains(t, err, tt.wantErrorSubstring)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantEndSession, a.endSessionEndpoint)
		})
	}
}
