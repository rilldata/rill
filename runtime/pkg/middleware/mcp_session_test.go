package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMCPSessionIDMiddleware(t *testing.T) {
	var seen string
	handler := MCPSessionIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("Mcp-Session-Id")
		w.WriteHeader(http.StatusOK)
	}))

	// A request without a session ID gets a new one, visible to the handler and returned to the client.
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/mcp", nil))
	issued := rec.Header().Get("Mcp-Session-Id")
	_, err := uuid.Parse(issued)
	require.NoError(t, err)
	require.Equal(t, issued, seen)

	// A request with a session ID keeps it.
	req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
	req.Header.Set("Mcp-Session-Id", "SESSION-123")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	require.Equal(t, "SESSION-123", seen)
	require.Equal(t, "SESSION-123", rec.Header().Get("Mcp-Session-Id"))

	// A malformed session ID is rejected before reaching the handler.
	for _, invalid := range []string{"session with spaces", "s\u00e9ssion", "session\x7f", strings.Repeat("s", maxMCPSessionIDLength+1)} {
		seen = ""
		req = httptest.NewRequest(http.MethodPost, "/mcp", nil)
		req.Header.Set("Mcp-Session-Id", invalid)
		rec = httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code, invalid)
		require.Empty(t, seen, invalid)
		require.Empty(t, rec.Header().Get("Mcp-Session-Id"), invalid)
	}
}
