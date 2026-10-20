package middleware

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// MCPSessionIDHeader is the header MCP clients use to carry a session ID.
// The MCP spec requires a client that receives it to send it on all its subsequent requests.
const MCPSessionIDHeader = "Mcp-Session-Id"

// maxMCPSessionIDLength bounds client-supplied session IDs, which the runtime stores as AI session IDs.
// Rill's own IDs are UUIDs, so the bound only matters for IDs that did not originate here.
const maxMCPSessionIDLength = 256

// MCPSessionIDMiddleware issues and preserves the Mcp-Session-Id header for a stateless MCP server.
// Since go-sdk v1.8, a stateless streamable HTTP handler neither reads nor sets the header,
// since the MCP spec is moving to sessionless servers (see SEP-2567).
// Rill nonetheless keys its AI sessions on the header, so that all the requests of one client connection share a session.
// A request without the header gets a new session ID, which is set on the request for the wrapped handler to read,
// and every response carries the session ID the request was served with.
// A request with a malformed session ID is rejected: the spec restricts session IDs to visible ASCII,
// and the length bound protects the catalog, which stores the ID as the AI session's ID.
func MCPSessionIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionID := r.Header.Get(MCPSessionIDHeader)
		if sessionID == "" {
			sessionID = uuid.NewString()
			r.Header.Set(MCPSessionIDHeader, sessionID)
		} else if len(sessionID) > maxMCPSessionIDLength || strings.ContainsFunc(sessionID, func(c rune) bool { return c < 0x21 || c > 0x7E }) {
			http.Error(w, "Bad Request: invalid Mcp-Session-Id header", http.StatusBadRequest)
			return
		}
		w.Header().Set(MCPSessionIDHeader, sessionID)
		next.ServeHTTP(w, r)
	})
}
