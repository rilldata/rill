package mcp

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime/drivers"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testMCPServer is an in-process MCP server with a few tools.
// If bearer tokens are configured, requests without an accepted token get a 401.
type testMCPServer struct {
	srv      *httptest.Server
	mu       sync.Mutex
	accepted map[string]bool
	requests int
	// sessions are the session IDs seen so far; expired ones get a 404, as after a server restart.
	sessions map[string]bool
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

// expireSessions makes the server forget all sessions seen so far.
func (s *testMCPServer) expireSessions() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id := range s.sessions {
		s.sessions[id] = false
	}
}

func newTestMCPServer(t *testing.T) *testMCPServer {
	t.Helper()
	ts := &testMCPServer{sessions: make(map[string]bool)}

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
		id := r.Header.Get("Mcp-Session-Id")
		live, seen := ts.sessions[id]
		if id != "" && !seen {
			ts.sessions[id] = true
		}
		ts.mu.Unlock()
		if seen && !live {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if accepted != nil {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token == "forbidden" {
				// A valid token that lacks access
				w.Header().Set("WWW-Authenticate", `Bearer error="insufficient_scope"`)
				w.WriteHeader(http.StatusForbidden)
				return
			}
			if !accepted[token] {
				w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer resource_metadata="%s/.well-known/oauth-protected-resource/mcp"`, ts.srv.URL))
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		handler.ServeHTTP(w, r)
	})
	ts.srv = httptest.NewServer(mux)
	t.Cleanup(ts.srv.Close)
	return ts
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

func TestReconnectAfterSessionExpired(t *testing.T) {
	ts := newTestMCPServer(t)
	h := openTestHandle(t, map[string]any{"url": ts.url()})
	ctx := t.Context()

	_, err := h.CallTool(ctx, "echo", map[string]any{"text": "hello"}, &drivers.MCPCallOptions{})
	require.NoError(t, err)

	ts.expireSessions()
	res, err := h.CallTool(ctx, "echo", map[string]any{"text": "again"}, &drivers.MCPCallOptions{})
	require.NoError(t, err)
	require.Equal(t, "again", res.Content)
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

	// Without a token, authorization is required
	_, err := h.ListTools(ctx, nil)
	require.ErrorIs(t, err, drivers.ErrMCPAuthorizationRequired)

	// The caller's token is used as-is in hosted mode
	tools, err := h.ListTools(ctx, &drivers.MCPCallOptions{OAuthToken: "user-token"})
	require.NoError(t, err)
	require.Len(t, tools, 4)

	_, err = h.CallTool(ctx, "echo", map[string]any{"text": "x"}, &drivers.MCPCallOptions{OAuthToken: "revoked"})
	require.ErrorIs(t, err, drivers.ErrMCPAuthorizationRequired)

	// A 403 means the token lacks access, which a new authorization can't fix, so it is not reported as authorization required
	_, err = h.CallTool(ctx, "echo", map[string]any{"text": "x"}, &drivers.MCPCallOptions{OAuthToken: "forbidden"})
	require.Error(t, err)
	require.NotErrorIs(t, err, drivers.ErrMCPAuthorizationRequired)
}

func TestOAuthConfigHash(t *testing.T) {
	base := &configProperties{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "secret", OAuthScopes: []string{"read"}}
	hash := oauthConfigHash(base)
	// Golden value: stored OAuth clients are keyed by this hash, so changing it for an unchanged config forces every user to reconnect
	require.Equal(t, "87eecbb220abb550eabd3dd68e691c92b97c53833d573360f612afea092c6e2e", hash)
	require.Equal(t, hash, oauthConfigHash(&configProperties{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "secret", OAuthScopes: []string{"read"}}))

	changes := []*configProperties{
		{URL: "https://other.example.com", OAuthClientID: "id", OAuthClientSecret: "secret", OAuthScopes: []string{"read"}},
		{URL: "https://mcp.example.com", OAuthClientID: "id2", OAuthClientSecret: "secret", OAuthScopes: []string{"read"}},
		{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "rotated", OAuthScopes: []string{"read"}},
		{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "secret", OAuthScopes: []string{"read", "write"}},
		{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "secret"},
		{URL: "https://mcp.example.com", OAuthClientID: "id", OAuthClientSecret: "secret", OAuthScopes: []string{"read"}, OAuthAuthorizationParams: map[string]string{"access_type": "offline"}},
		// Shifting a value between fields must not collide
		{URL: "https://mcp.example.com", OAuthClientID: "idsecret", OAuthScopes: []string{"read"}},
	}
	for _, c := range changes {
		require.NotEqual(t, hash, oauthConfigHash(c))
	}

	// Changing the value of an authorization param changes the hash
	require.NotEqual(t,
		oauthConfigHash(&configProperties{URL: "https://mcp.example.com", OAuthAuthorizationParams: map[string]string{"access_type": "offline"}}),
		oauthConfigHash(&configProperties{URL: "https://mcp.example.com", OAuthAuthorizationParams: map[string]string{"access_type": "online"}}),
	)

	// Authorization params must not collide with scopes
	require.NotEqual(t,
		oauthConfigHash(&configProperties{URL: "https://mcp.example.com", OAuthScopes: []string{"access_type", "offline"}}),
		oauthConfigHash(&configProperties{URL: "https://mcp.example.com", OAuthAuthorizationParams: map[string]string{"access_type": "offline"}}),
	)

	// Authorization params are hashed independently of their order
	params := map[string]string{"access_type": "offline", "prompt": "consent", "hd": "example.com"}
	withParams := &configProperties{URL: "https://mcp.example.com", OAuthAuthorizationParams: params}
	for range 10 {
		require.Equal(t, oauthConfigHash(withParams), oauthConfigHash(&configProperties{URL: "https://mcp.example.com", OAuthAuthorizationParams: maps.Clone(params)}))
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
		{"reserved authorization param", map[string]any{"url": "https://example.com/mcp", "auth": "oauth", "oauth_authorization_params": map[string]any{"State": "x"}}, "cannot set \"State\""},
		{"authorization params without oauth", map[string]any{"url": "https://example.com/mcp", "oauth_authorization_params": map[string]any{"prompt": "consent"}}, "requires 'auth: oauth'"},
		{"bad timeout", map[string]any{"url": "https://example.com/mcp", "timeout": "soon"}, "invalid 'timeout'"},
		{"ok hosted", map[string]any{"url": "https://example.com/mcp", "allow_host_access": false, "timeout": "30s", "max_result_bytes": 10}, ""},
		{"ok authorization params", map[string]any{"url": "https://example.com/mcp", "auth": "oauth", "oauth_authorization_params": map[string]any{"access_type": "offline", "prompt": "consent"}}, ""},
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

func TestTruncateUTF8(t *testing.T) {
	require.Equal(t, "héllo", truncateUTF8("héllo", 10))
	// "é" is two bytes; cutting inside it must back off to the previous character
	require.Equal(t, "h", truncateUTF8("héllo", 2))
	require.Equal(t, "hé", truncateUTF8("héllo", 3))
	require.Equal(t, "", truncateUTF8("é", 1))
	require.Equal(t, "日本", truncateUTF8("日本語", 7))
}
