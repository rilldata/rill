package mcp

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/pkg/oauthclient"
	"github.com/rilldata/rill/runtime/pkg/safehttp"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
)

// testMCPServer is an in-process MCP server with a few tools.
// If bearer tokens are configured, requests without an accepted token get a 401 with protected resource metadata.
type testMCPServer struct {
	srv      *httptest.Server
	asURL    string
	mu       sync.Mutex
	accepted map[string]bool
	requests int
}

func (s *testMCPServer) url() string {
	return s.srv.URL + "/mcp"
}

func (s *testMCPServer) accept(tokens ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accepted = make(map[string]bool)
	for _, t := range tokens {
		s.accepted[t] = true
	}
}

func newTestMCPServer(t *testing.T) *testMCPServer {
	t.Helper()
	ts := &testMCPServer{}

	server := gomcp.NewServer(&gomcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	type echoArgs struct {
		Text string `json:"text"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "echo", Description: "Echoes the text"}, func(ctx context.Context, req *gomcp.CallToolRequest, args echoArgs) (*gomcp.CallToolResult, any, error) {
		return &gomcp.CallToolResult{Content: []gomcp.Content{&gomcp.TextContent{Text: args.Text}}}, nil, nil
	})
	gomcp.AddTool(server, &gomcp.Tool{Name: "whoami", Description: "Returns the request headers"}, func(ctx context.Context, req *gomcp.CallToolRequest, args struct{}) (*gomcp.CallToolResult, any, error) {
		text := req.Extra.Header.Get("X-User") + "/" + req.Extra.Header.Get("X-Env")
		return &gomcp.CallToolResult{Content: []gomcp.Content{&gomcp.TextContent{Text: text}}}, nil, nil
	})
	gomcp.AddTool(server, &gomcp.Tool{Name: "big", Description: "Returns a large result"}, func(ctx context.Context, req *gomcp.CallToolRequest, args struct{}) (*gomcp.CallToolResult, any, error) {
		return &gomcp.CallToolResult{Content: []gomcp.Content{&gomcp.TextContent{Text: strings.Repeat("x", 300*1024)}}}, nil, nil
	})
	gomcp.AddTool(server, &gomcp.Tool{Name: "fail", Description: "Always fails"}, func(ctx context.Context, req *gomcp.CallToolRequest, args struct{}) (*gomcp.CallToolResult, any, error) {
		return &gomcp.CallToolResult{IsError: true, Content: []gomcp.Content{&gomcp.TextContent{Text: "boom"}}}, nil, nil
	})
	handler := gomcp.NewStreamableHTTPHandler(func(*http.Request) *gomcp.Server { return server }, nil)

	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		ts.mu.Lock()
		ts.requests++
		accepted := ts.accepted
		ts.mu.Unlock()
		if accepted != nil {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !accepted[token] {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer resource_metadata="%s/.well-known/oauth-protected-resource/mcp"`, ts.srv.URL))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		handler.ServeHTTP(w, r)
	})
	mux.HandleFunc("/.well-known/oauth-protected-resource/mcp", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"resource":              ts.srv.URL + "/mcp",
			"authorization_servers": []string{ts.asURL},
			"scopes_supported":      []string{"read"},
		})
	})
	ts.srv = httptest.NewServer(mux)
	t.Cleanup(ts.srv.Close)
	return ts
}

// testAuthServer is a fake OAuth 2.1 authorization server with dynamic client registration, PKCE and refresh token rotation.
type testAuthServer struct {
	srv             *httptest.Server
	mu              sync.Mutex
	challenge       string // expected S256 code challenge, taken from the authorization URL by the test
	resource        string // expected resource indicator
	issued          int
	refreshes       int
	lastRefresh     string
	revoked         []string
	revokedClientID string
	registerCall    int
}

func newTestAuthServer(t *testing.T) *testAuthServer {
	t.Helper()
	as := &testAuthServer{}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                as.srv.URL,
			"authorization_endpoint":                as.srv.URL + "/authorize",
			"token_endpoint":                        as.srv.URL + "/token",
			"registration_endpoint":                 as.srv.URL + "/register",
			"revocation_endpoint":                   as.srv.URL + "/revoke",
			"response_types_supported":              []string{"code"},
			"grant_types_supported":                 []string{"authorization_code", "refresh_token"},
			"code_challenge_methods_supported":      []string{"S256"},
			"token_endpoint_auth_methods_supported": []string{"none", "client_secret_post"},
		})
	})
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		as.mu.Lock()
		as.registerCall++
		as.mu.Unlock()
		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		req["client_id"] = "client-123"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(req)
	})
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		as.mu.Lock()
		defer as.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		if r.Form.Get("resource") != as.resource {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_target", "error_description": "unexpected resource " + r.Form.Get("resource")})
			return
		}
		if r.Form.Get("client_id") != "client-123" {
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_client"})
			return
		}
		switch r.Form.Get("grant_type") {
		case "authorization_code":
			sum := sha256.Sum256([]byte(r.Form.Get("code_verifier")))
			if r.Form.Get("code") != "code123" || base64.RawURLEncoding.EncodeToString(sum[:]) != as.challenge {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_grant"})
				return
			}
		case "refresh_token":
			as.refreshes++
			as.lastRefresh = r.Form.Get("refresh_token")
			if as.lastRefresh != fmt.Sprintf("rt%d", as.issued) {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]any{"error": "invalid_grant"})
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{"error": "unsupported_grant_type"})
			return
		}
		as.issued++
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token":  fmt.Sprintf("at%d", as.issued),
			"token_type":    "bearer",
			"refresh_token": fmt.Sprintf("rt%d", as.issued),
			"expires_in":    3600,
		})
	})
	mux.HandleFunc("/revoke", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		as.mu.Lock()
		as.revoked = append(as.revoked, r.Form.Get("token"))
		as.revokedClientID = r.Form.Get("client_id")
		as.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	})
	as.srv = httptest.NewServer(mux)
	t.Cleanup(as.srv.Close)
	return as
}

func openTestHandle(t *testing.T, config map[string]any) *handle {
	t.Helper()
	if _, ok := config["allow_host_access"]; !ok {
		config["allow_host_access"] = true
	}
	st := storage.MustNew(t.TempDir(), nil).WithPrefix("inst", "conn")
	h, err := driver{}.Open(t.Context(), "conn", "inst", config, st, activity.NewNoopClient(), zap.NewNop())
	require.NoError(t, err)
	t.Cleanup(func() { _ = h.Close() })
	return h.(*handle)
}

func TestListAndCallTools(t *testing.T) {
	ts := newTestMCPServer(t)
	h := openTestHandle(t, map[string]any{
		"url": ts.url(),
		"headers": map[string]any{
			"X-User": "{{ .user.email }}",
			"X-Env":  "{{ .env.secret }}",
		},
	})
	ctx := t.Context()
	opts := &drivers.MCPCallOptions{
		Variables:      map[string]string{"secret": "s3cret"},
		UserAttributes: map[string]any{"email": "a@example.com"},
	}

	require.NoError(t, h.Ping(ctx))

	tools, err := h.ListTools(ctx, opts)
	require.NoError(t, err)
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
		require.NotEmpty(t, tool.Description)
		require.NotNil(t, tool.InputSchema)
	}
	require.ElementsMatch(t, []string{"echo", "whoami", "big", "fail"}, names)

	// Listing again is served from the cache
	before := ts.requests
	_, err = h.ListTools(ctx, opts)
	require.NoError(t, err)
	require.Equal(t, before, ts.requests)

	res, err := h.CallTool(ctx, "echo", map[string]any{"text": "hello"}, opts)
	require.NoError(t, err)
	require.Equal(t, "hello", res.Content)
	require.False(t, res.IsError)

	// Headers are resolved per user and per environment
	res, err = h.CallTool(ctx, "whoami", nil, opts)
	require.NoError(t, err)
	require.Equal(t, "a@example.com/s3cret", res.Content)

	res, err = h.CallTool(ctx, "whoami", nil, &drivers.MCPCallOptions{
		Variables:      map[string]string{"secret": "s3cret"},
		UserAttributes: map[string]any{"email": "b@example.com"},
	})
	require.NoError(t, err)
	require.Equal(t, "b@example.com/s3cret", res.Content)

	// A missing user attribute is an error rather than an empty header
	_, err = h.CallTool(ctx, "whoami", nil, &drivers.MCPCallOptions{Variables: map[string]string{"secret": "s3cret"}, UserAttributes: map[string]any{}})
	require.ErrorContains(t, err, "X-User")

	// Callers without a user (the reconciler) cannot use a server whose headers identify the user
	_, err = h.ListTools(ctx, &drivers.MCPCallOptions{Variables: map[string]string{"secret": "s3cret"}})
	require.ErrorIs(t, err, drivers.ErrMCPAuthorizationRequired)

	// Tool errors are results, not errors
	res, err = h.CallTool(ctx, "fail", nil, opts)
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.Equal(t, "boom", res.Content)

	// Large results are truncated to the configured limit
	res, err = h.CallTool(ctx, "big", nil, opts)
	require.NoError(t, err)
	require.True(t, res.Truncated)
	require.Less(t, len(res.Content), 300*1024)
	require.Contains(t, res.Content, "[truncated")
}

func TestStaticBearerToken(t *testing.T) {
	ts := newTestMCPServer(t)
	ts.accept("static-token")

	h := openTestHandle(t, map[string]any{
		"url":     ts.url(),
		"headers": map[string]any{"Authorization": "Bearer {{ .env.token }}"},
	})
	ctx := t.Context()

	_, err := h.ListTools(ctx, &drivers.MCPCallOptions{Variables: map[string]string{"token": "static-token"}})
	require.NoError(t, err)

	_, err = h.ListTools(ctx, &drivers.MCPCallOptions{Variables: map[string]string{"token": "wrong"}})
	require.ErrorIs(t, err, drivers.ErrMCPAuthorizationRequired)

	// Ping only checks reachability, so a 401 is fine
	require.NoError(t, h.Ping(ctx))
}

func TestHostedOAuthToken(t *testing.T) {
	ts := newTestMCPServer(t)
	ts.accept("user-token")

	h := openTestHandle(t, map[string]any{"url": ts.url(), "auth": "oauth"})
	ctx := t.Context()

	// Without a token (and no local state), authorization is required
	_, err := h.ListTools(ctx, nil)
	require.ErrorIs(t, err, drivers.ErrMCPAuthorizationRequired)

	// The caller's token is used as-is in hosted mode
	tools, err := h.ListTools(ctx, &drivers.MCPCallOptions{OAuthToken: "user-token"})
	require.NoError(t, err)
	require.Len(t, tools, 4)

	_, err = h.CallTool(ctx, "echo", map[string]any{"text": "x"}, &drivers.MCPCallOptions{OAuthToken: "revoked"})
	require.ErrorIs(t, err, drivers.ErrMCPAuthorizationRequired)
}

func TestLocalOAuthFlow(t *testing.T) {
	as := newTestAuthServer(t)
	ts := newTestMCPServer(t)
	ts.asURL = as.srv.URL
	ts.accept() // nothing accepted until tokens are issued

	h := openTestHandle(t, map[string]any{"url": ts.url(), "auth": "oauth", "oauth_scopes": []any{"read"}})
	ctx := t.Context()
	var manager OAuthManager = h
	require.True(t, manager.OAuthEnabled())
	cfg := h.OAuthConfig()
	require.Equal(t, ts.url(), cfg.ServerURL)
	require.Equal(t, []string{"read"}, cfg.Scopes)
	require.NotEmpty(t, cfg.Hash)

	connected, _, err := manager.OAuthStatus(ctx)
	require.NoError(t, err)
	require.False(t, connected)
	authorized, err := h.OAuthAuthorized(ctx)
	require.NoError(t, err)
	require.False(t, authorized)

	// Begin: discovers the authorization server through the well-known metadata and registers a client
	authURL, verifier, err := manager.BeginOAuth(ctx, "http://localhost:9009/mcp/oauth/callback", "state-1")
	require.NoError(t, err)
	u, err := url.Parse(authURL)
	require.NoError(t, err)
	require.Equal(t, as.srv.URL+"/authorize", u.Scheme+"://"+u.Host+u.Path)
	q := u.Query()
	require.Equal(t, "state-1", q.Get("state"))
	require.Equal(t, "client-123", q.Get("client_id"))
	require.Equal(t, "S256", q.Get("code_challenge_method"))
	require.Equal(t, ts.url(), q.Get("resource"))
	require.Equal(t, "read", q.Get("scope"))
	require.Equal(t, "http://localhost:9009/mcp/oauth/callback", q.Get("redirect_uri"))
	require.Equal(t, 1, as.registerCall)

	// The client is persisted before the token so the callback can use it
	statePath := filepath.Join(mustDataDir(t, h), oauthStateFilename)
	info, err := os.Stat(statePath)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), info.Mode().Perm())

	// Complete: exchanges the code with PKCE and the resource indicator
	as.challenge = q.Get("code_challenge")
	as.resource = ts.url()
	require.ErrorContains(t, manager.CompleteOAuth(ctx, "code123", verifier, "https://evil.example.com"), "not the expected authorization server")
	require.NoError(t, manager.CompleteOAuth(ctx, "code123", verifier, ""))
	ts.accept("at1", "at2", "at3")

	connected, expiresOn, err := manager.OAuthStatus(ctx)
	require.NoError(t, err)
	require.True(t, connected)
	require.WithinDuration(t, time.Now().Add(time.Hour), expiresOn, time.Minute)
	authorized, err = h.OAuthAuthorized(ctx)
	require.NoError(t, err)
	require.True(t, authorized)

	// Calls use the persisted token
	res, err := h.CallTool(ctx, "echo", map[string]any{"text": "hi"}, nil)
	require.NoError(t, err)
	require.Equal(t, "hi", res.Content)

	// An expiring token is refreshed and the rotated refresh token is persisted
	h.oauth.mu.Lock()
	h.oauth.state.Token.Expiry = time.Now().Add(-time.Minute)
	h.oauth.mu.Unlock()
	res, err = h.CallTool(ctx, "echo", map[string]any{"text": "again"}, nil)
	require.NoError(t, err)
	require.Equal(t, "again", res.Content)
	require.Equal(t, 1, as.refreshes)
	require.Equal(t, "rt1", as.lastRefresh)
	persisted := readOAuthState(t, statePath)
	require.Equal(t, "at2", persisted.Token.AccessToken)
	require.Equal(t, "rt2", persisted.Token.RefreshToken)

	// A token the server rejects is refreshed once and the call retried
	ts.accept("at3")
	res, err = h.CallTool(ctx, "echo", map[string]any{"text": "third"}, nil)
	require.NoError(t, err)
	require.Equal(t, "third", res.Content)
	require.Equal(t, 2, as.refreshes)
	require.Equal(t, "at3", readOAuthState(t, statePath).Token.AccessToken)

	// Beginning a new flow reuses the registered client
	_, _, err = manager.BeginOAuth(ctx, "http://localhost:9009/mcp/oauth/callback", "state-2")
	require.NoError(t, err)
	require.Equal(t, 1, as.registerCall)

	// State persisted for another OAuth configuration is ignored
	h.oauth.mu.Lock()
	h.oauth.loaded = false
	h.oauthConfigHash = "other"
	h.oauth.mu.Unlock()
	authorized, err = h.OAuthAuthorized(ctx)
	require.NoError(t, err)
	require.False(t, authorized)

	// Disconnecting forgets the token
	require.NoError(t, h.ResetOAuth(ctx))
	_, err = os.Stat(statePath)
	require.True(t, os.IsNotExist(err))
	_, err = h.ListTools(ctx, nil)
	require.ErrorIs(t, err, drivers.ErrMCPAuthorizationRequired)
}

func TestLocalTokenWithoutRefreshToken(t *testing.T) {
	serverURL := "https://example.com/mcp"
	h := openTestHandle(t, map[string]any{"url": serverURL, "auth": "oauth"})
	ctx := t.Context()

	// Within the refresh window, a token that can't be refreshed is still used until it expires
	require.NoError(t, h.oauth.save(&OAuthState{
		Client:          &oauthclient.Client{ServerURL: serverURL},
		Token:           &oauth2.Token{AccessToken: "at1", Expiry: time.Now().Add(time.Minute)},
		TokenObtainedOn: time.Now().Add(-time.Hour),
	}))
	tok, err := h.oauth.accessToken(ctx, false)
	require.NoError(t, err)
	require.Equal(t, "at1", tok)

	// Once expired, the user must authorize again
	require.NoError(t, h.oauth.save(&OAuthState{
		Client:          &oauthclient.Client{ServerURL: serverURL},
		Token:           &oauth2.Token{AccessToken: "at1", Expiry: time.Now().Add(-time.Minute)},
		TokenObtainedOn: time.Now().Add(-time.Hour),
	}))
	_, err = h.oauth.accessToken(ctx, false)
	require.ErrorIs(t, err, drivers.ErrMCPAuthorizationRequired)
}

func TestPreregisteredOAuthClient(t *testing.T) {
	as := newTestAuthServer(t)
	ts := newTestMCPServer(t)
	ts.asURL = as.srv.URL

	client, err := oauthclient.Discover(t.Context(), &oauthclient.DiscoverOptions{
		ServerURL:    ts.url(),
		RedirectURL:  "http://localhost:9009/mcp/oauth/callback",
		ClientID:     "pre-registered",
		ClientSecret: "shh",
		HTTPClient:   safehttp.NewHTTPClient(true, 5*time.Second),
	})
	require.NoError(t, err)
	require.Equal(t, "pre-registered", client.ClientID)
	require.Equal(t, "shh", client.ClientSecret)
	require.Equal(t, 0, as.registerCall)
	require.Equal(t, []string{"read"}, client.Scopes) // defaults to the scopes the server advertises
	require.Equal(t, as.srv.URL+"/revoke", client.RevocationEndpoint)
}

func TestOAuthConfigHash(t *testing.T) {
	base := &configProperties{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "secret", OAuthScopes: []string{"read"}}
	hash := oauthConfigHash(base)
	require.Equal(t, hash, oauthConfigHash(&configProperties{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "secret", OAuthScopes: []string{"read"}}))

	changes := []*configProperties{
		{URL: "https://other.example.com", OAuthClientID: "id", OAuthClientSecret: "secret", OAuthScopes: []string{"read"}},
		{URL: "https://mcp.example.com", OAuthClientID: "id2", OAuthClientSecret: "secret", OAuthScopes: []string{"read"}},
		{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "rotated", OAuthScopes: []string{"read"}},
		{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "secret", OAuthScopes: []string{"read", "write"}},
		{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "secret"},
		// Shifting a value between fields must not collide
		{URL: "https://mcp.example.com", OAuthClientID: "idsecret", OAuthScopes: []string{"read"}},
	}
	for _, c := range changes {
		require.NotEqual(t, hash, oauthConfigHash(c))
	}
}

func TestConfigValidation(t *testing.T) {
	cases := []struct {
		name   string
		config map[string]any
		err    string
	}{
		{"missing url", map[string]any{}, "missing required property 'url'"},
		{"bad scheme", map[string]any{"url": "ftp://x/mcp"}, "scheme must be http or https"},
		{"http in hosted", map[string]any{"url": "http://example.com/mcp", "allow_host_access": false}, "must use https"},
		{"localhost in hosted", map[string]any{"url": "https://localhost/mcp", "allow_host_access": false}, "private networks"},
		{"private ip in hosted", map[string]any{"url": "https://10.1.2.3/mcp", "allow_host_access": false}, "private networks"},
		{"bad auth", map[string]any{"url": "https://example.com/mcp", "auth": "basic"}, "invalid 'auth'"},
		{"oauth with authorization header", map[string]any{"url": "https://example.com/mcp", "auth": "oauth", "headers": map[string]any{"Authorization": "Bearer x"}}, "cannot be combined"},
		{"bad timeout", map[string]any{"url": "https://example.com/mcp", "timeout": "soon"}, "invalid 'timeout'"},
		{"ok hosted", map[string]any{"url": "https://example.com/mcp", "allow_host_access": false, "timeout": "30s", "max_result_bytes": 10}, ""},
		{"ok local http", map[string]any{"url": "http://localhost:8000/mcp", "allow_host_access": true}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := newConfigProperties(tc.config)
			if tc.err == "" {
				require.NoError(t, err)
				require.NotNil(t, cfg)
				return
			}
			require.ErrorContains(t, err, tc.err)
		})
	}
}

func mustDataDir(t *testing.T, h *handle) string {
	dir, err := h.storage.DataDir()
	require.NoError(t, err)
	return dir
}

func readOAuthState(t *testing.T, path string) *OAuthState {
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	state := &OAuthState{}
	require.NoError(t, json.Unmarshal(data, state))
	return state
}

func TestTruncateUTF8(t *testing.T) {
	require.Equal(t, "héllo", truncateUTF8("héllo", 10))
	// "é" is two bytes; cutting inside it must back off to the previous character
	require.Equal(t, "h", truncateUTF8("héllo", 2))
	require.Equal(t, "hé", truncateUTF8("héllo", 3))
	require.Equal(t, "", truncateUTF8("é", 1))
	require.Equal(t, "日本", truncateUTF8("日本語", 7))
}
