package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

// mcpSessionIDHeader is the header MCP clients use to carry a session ID.
// The MCP spec requires a client that receives it to send it on all its subsequent requests.
const mcpSessionIDHeader = "Mcp-Session-Id"

// MCPSessionIDMiddleware issues and preserves the Mcp-Session-Id header for a stateless MCP server.
// Since go-sdk v1.8, a stateless streamable HTTP handler neither reads nor sets the header,
// since the MCP spec is moving to sessionless servers (see SEP-2567).
// Rill nonetheless keys its AI sessions on the header, so that all the requests of one client connection share a session.
// A request without the header gets a new session ID, which is set on the request for the wrapped handler to read,
// and every response carries the session ID the request was served with.
func MCPSessionIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get(mcpSessionIDHeader)
		if sessionID == "" {
			sessionID = uuid.NewString()
			r.Header.Set(mcpSessionIDHeader, sessionID)
		}
		w.Header().Set(mcpSessionIDHeader, sessionID)
		next.ServeHTTP(w, r)
	})
}
