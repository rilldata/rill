package mcp

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/rilldata/rill/runtime/pkg/activity"
	"github.com/rilldata/rill/runtime/storage"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// testMCPServer is an in-process MCP server with a few tools.
// If a token is set, requests without it get a 401.
type testMCPServer struct {
	srv      *httptest.Server
	token    string
	mu       sync.Mutex
	requests int
}

func (s *testMCPServer) url() string {
	return s.srv.URL + "/mcp"
}

func newTestMCPServer(t *testing.T, token string) *testMCPServer {
	t.Helper()
	ts := &testMCPServer{token: token}

	server := gomcp.NewServer(&gomcp.Implementation{Name: "test", Version: "1.0.0"}, nil)
	type echoArgs struct {
		Text string `json:"text"`
	}
	gomcp.AddTool(server, &gomcp.Tool{Name: "echo", Description: "Echoes the text"}, func(ctx context.Context, req *gomcp.CallToolRequest, args echoArgs) (*gomcp.CallToolResult, any, error) {
		return &gomcp.CallToolResult{Content: []gomcp.Content{&gomcp.TextContent{Text: args.Text}}}, nil, nil
	})
	gomcp.AddTool(server, &gomcp.Tool{Name: "header", Description: "Returns the X-Custom request header"}, func(ctx context.Context, req *gomcp.CallToolRequest, args struct{}) (*gomcp.CallToolResult, any, error) {
		return &gomcp.CallToolResult{Content: []gomcp.Content{&gomcp.TextContent{Text: req.Extra.Header.Get("X-Custom")}}}, nil, nil
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
		ts.mu.Unlock()
		if ts.token != "" && r.Header.Get("Authorization") != "Bearer "+ts.token {
			w.WriteHeader(http.StatusUnauthorized)
			return
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
	ts := newTestMCPServer(t, "")
	h := openTestHandle(t, map[string]any{
		"url":     ts.url(),
		"headers": map[string]any{"X-Custom": "value"},
	})
	ctx := t.Context()

	require.NoError(t, h.Ping(ctx))

	tools, err := h.ListTools(ctx)
	require.NoError(t, err)
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
		require.NotEmpty(t, tool.Description)
		require.NotNil(t, tool.InputSchema)
	}
	require.ElementsMatch(t, []string{"echo", "header", "big", "fail"}, names)

	// Listing again is served from the cache
	ts.mu.Lock()
	before := ts.requests
	ts.mu.Unlock()
	_, err = h.ListTools(ctx)
	require.NoError(t, err)
	ts.mu.Lock()
	require.Equal(t, before, ts.requests)
	ts.mu.Unlock()

	res, err := h.CallTool(ctx, "echo", map[string]any{"text": "hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", res.Content)
	require.False(t, res.IsError)

	// The connector's headers are sent on every request
	res, err = h.CallTool(ctx, "header", nil)
	require.NoError(t, err)
	require.Equal(t, "value", res.Content)

	// Tool errors are results, not errors
	res, err = h.CallTool(ctx, "fail", nil)
	require.NoError(t, err)
	require.True(t, res.IsError)
	require.Equal(t, "boom", res.Content)

	// Large results are truncated to the configured limit
	res, err = h.CallTool(ctx, "big", nil)
	require.NoError(t, err)
	require.True(t, res.Truncated)
	require.Less(t, len(res.Content), 300*1024)
	require.Contains(t, res.Content, "[truncated")
}

func TestStaticBearerToken(t *testing.T) {
	ts := newTestMCPServer(t, "static-token")
	ctx := t.Context()

	h := openTestHandle(t, map[string]any{
		"url":     ts.url(),
		"headers": map[string]any{"Authorization": "Bearer static-token"},
	})
	require.NoError(t, h.Ping(ctx))
	_, err := h.ListTools(ctx)
	require.NoError(t, err)

	// Ping connects, so it reports rejected credentials
	h = openTestHandle(t, map[string]any{
		"url":     ts.url(),
		"headers": map[string]any{"Authorization": "Bearer wrong"},
	})
	require.Error(t, h.Ping(ctx))
	_, err = h.ListTools(ctx)
	require.Error(t, err)
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

func TestSafeDialRejectsPrivateAddresses(t *testing.T) {
	ts := newTestMCPServer(t, "")
	// The test server listens on 127.0.0.1, which hosted runtimes must never reach, even through a hostname.
	hosted := newTransport(false, time.Second)
	_, err := hosted.DialContext(t.Context(), "tcp", strings.TrimPrefix(ts.srv.URL, "http://"))
	require.ErrorContains(t, err, "private address")

	// Rill Developer may reach local servers
	local := newTransport(true, time.Second)
	conn, err := local.DialContext(t.Context(), "tcp", strings.TrimPrefix(ts.srv.URL, "http://"))
	require.NoError(t, err)
	require.NoError(t, conn.Close())

	require.True(t, isPrivateIP(net.ParseIP("169.254.169.254")))
	require.True(t, isPrivateIP(net.ParseIP("::1")))
	require.True(t, isPrivateIP(net.ParseIP("fd00::1")))
	require.False(t, isPrivateIP(net.ParseIP("93.184.216.34")))
}

func TestTruncateUTF8(t *testing.T) {
	require.Equal(t, "héllo", truncateUTF8("héllo", 10))
	// "é" is two bytes; cutting inside it must back off to the previous character
	require.Equal(t, "h", truncateUTF8("héllo", 2))
	require.Equal(t, "hé", truncateUTF8("héllo", 3))
	require.Equal(t, "", truncateUTF8("é", 1))
	require.Equal(t, "日本", truncateUTF8("日本語", 7))
}
