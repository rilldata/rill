package middleware

import (
	"net/http"
	"net/http/httptest"
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
}
