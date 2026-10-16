package oauthclient

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/rilldata/rill/runtime/pkg/safehttp"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

// newDiscoveryServer serves authorization server metadata on its origin, and protected resource metadata for /mcp if withPRM is set.
func newDiscoveryServer(t *testing.T, withPRM bool) *httptest.Server {
	var srv *httptest.Server
	mux := http.NewServeMux()
	if withPRM {
		mux.HandleFunc("/.well-known/oauth-protected-resource/mcp", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"resource":              srv.URL + "/mcp",
				"authorization_servers": []string{srv.URL},
				"scopes_supported":      []string{"read", "write"},
			})
		})
	}
	mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                           srv.URL,
			"authorization_endpoint":           srv.URL + "/authorize",
			"token_endpoint":                   srv.URL + "/token",
			"response_types_supported":         []string{"code"},
			"code_challenge_methods_supported": []string{"S256"},
		})
	})
	srv = httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func TestDiscover(t *testing.T) {
	discover := func(serverURL string, scopes []string) (*Client, error) {
		return Discover(t.Context(), &DiscoverOptions{
			ServerURL:   serverURL,
			RedirectURL: "http://localhost:9009/callback",
			ClientID:    "pre-registered",
			Scopes:      scopes,
			HTTPClient:  safehttp.NewHTTPClient(true, 5*time.Second),
		})
	}

	t.Run("protected resource metadata", func(t *testing.T) {
		srv := newDiscoveryServer(t, true)
		c, err := discover(srv.URL+"/mcp", nil)
		require.NoError(t, err)
		require.Equal(t, srv.URL, c.Issuer)
		require.Equal(t, srv.URL+"/mcp", c.Resource)
		require.Equal(t, srv.URL+"/authorize", c.AuthorizationEndpoint)
		require.Equal(t, []string{"read", "write"}, c.Scopes)
	})

	t.Run("trailing slash matches the declared resource", func(t *testing.T) {
		srv := newDiscoveryServer(t, true)
		c, err := discover(srv.URL+"/mcp/", []string{"read"})
		require.NoError(t, err)
		require.Equal(t, srv.URL+"/mcp", c.Resource)
		require.Equal(t, []string{"read"}, c.Scopes)
	})

	t.Run("falls back to authorization server metadata on the origin", func(t *testing.T) {
		srv := newDiscoveryServer(t, false)
		c, err := discover(srv.URL+"/mcp", nil)
		require.NoError(t, err)
		require.Equal(t, srv.URL, c.Issuer)
		require.Equal(t, srv.URL+"/mcp", c.Resource)
		require.Equal(t, srv.URL+"/token", c.TokenEndpoint)
		require.Empty(t, c.Scopes)
	})

	t.Run("fails without any metadata", func(t *testing.T) {
		srv := httptest.NewServer(http.NotFoundHandler())
		t.Cleanup(srv.Close)
		_, err := discover(srv.URL+"/mcp", nil)
		require.ErrorContains(t, err, "serves no authorization server metadata")
	})
}

func TestValidateIssuer(t *testing.T) {
	cases := []struct {
		name      string
		supported bool
		iss       string
		ok        bool
	}{
		{"advertised and matching", true, "https://as.example.com", true},
		{"advertised and missing", true, "", false},
		{"advertised and mismatched", true, "https://evil.example.com", false},
		{"unadvertised and missing", false, "", true},
		{"unadvertised and matching", false, "https://as.example.com", true},
		{"unadvertised and mismatched", false, "https://evil.example.com", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &Client{Issuer: "https://as.example.com", IssParameterSupported: tc.supported}
			err := c.ValidateIssuer(tc.iss)
			if tc.ok {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestTokenExpiring(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name       string
		expiresIn  time.Duration
		obtainedOn time.Time
		expiring   bool
	}{
		{"no expiry", 0, now, false},
		{"long-lived, far from expiry", 30 * time.Minute, now.Add(-30 * time.Minute), false},
		{"long-lived, within max leeway", 4 * time.Minute, now.Add(-56 * time.Minute), true},
		{"short-lived, within half its lifetime", 2 * time.Minute, now.Add(-4 * time.Minute), true},
		{"short-lived, fresh", 5 * time.Minute, now.Add(-time.Minute), false},
		{"expired", -time.Minute, now.Add(-time.Hour), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tok := &oauth2.Token{AccessToken: "x"}
			if tc.expiresIn != 0 {
				tok.Expiry = now.Add(tc.expiresIn)
			}
			require.Equal(t, tc.expiring, TokenExpiring(tok, tc.obtainedOn))
		})
	}
}

func TestIsLinkableRedirect(t *testing.T) {
	require.True(t, isLinkableRedirect("/org/project"))
	require.True(t, isLinkableRedirect("https://ui.rilldata.com/org/project"))
	require.True(t, isLinkableRedirect("http://localhost:9009/"))
	require.False(t, isLinkableRedirect(""))
	require.False(t, isLinkableRedirect("javascript:alert(1)"))
	require.False(t, isLinkableRedirect("data:text/html,hi"))
	require.False(t, isLinkableRedirect("//evil.com/x"))
	require.False(t, isLinkableRedirect("ftp://example.com/"))
}

func TestRefreshKeepsGrantedScopes(t *testing.T) {
	var form url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		form = r.Form
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"access_token": "at2", "token_type": "bearer", "refresh_token": "rt2", "expires_in": 3600})
	}))
	t.Cleanup(srv.Close)

	c := &Client{TokenEndpoint: srv.URL + "/token", ClientID: "client-123", Resource: "https://mcp.example.com/mcp", Scopes: []string{"read", "write"}}
	tok, err := c.Refresh(t.Context(), &oauth2.Token{AccessToken: "at1", RefreshToken: "rt1"}, safehttp.NewHTTPClient(true, 5*time.Second))
	require.NoError(t, err)
	require.Equal(t, "at2", tok.AccessToken)
	require.Equal(t, "rt2", tok.RefreshToken)
	require.Equal(t, "refresh_token", form.Get("grant_type"))
	require.Equal(t, "rt1", form.Get("refresh_token"))
	require.Equal(t, c.Resource, form.Get("resource"))
	// The scope is omitted so the new token keeps the scopes of the original grant, which may be fewer than configured.
	require.Empty(t, form.Get("scope"))
}
