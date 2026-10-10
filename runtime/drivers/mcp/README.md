# mcp

Connector driver for external MCP (Model Context Protocol) servers. It lets Rill's AI agents discover and call tools on a remote server over streamable HTTP.

## Files

- `mcp.go`: driver registration, config validation, the handle with its per-credential session and tool caches, and result flattening.
- `oauth.go`: OAuth 2.1 client support (discovery, dynamic client registration, PKCE, resource indicators, refresh with rotation), the OAuth state Rill Developer persists locally, and the `OAuthManager` interface the local server uses to run the browser flow.
- `transport.go`: the HTTP transport, including the dialer that refuses private addresses in hosted deployments and the round tripper that injects credentials and records authorization failures.

## How it fits together

- The parser (`runtime/parser/parse_connector.go`) validates the connector, lifts its `security` block into security rules on the connector spec, and leaves header templates unresolved because they may reference the requesting user.
- The connector reconciler pings the server and lists its tools into the connector state with the project's credentials. Servers that need a per-user authorization (OAuth, or headers that reference the user) are only checked for reachability.
- The AI layer (`runtime/ai/external_tools.go`) discovers servers per session, applies the security policy through the security engine, compiles the tools for the agents, and resolves per-user OAuth tokens through the admin service in Rill Cloud.
- Rill Developer's local server (`cli/pkg/local/mcp_oauth.go`) and the admin service (`admin/server/mcp_oauth.go`) run the browser-based authorization flow using the helpers in `oauth.go`.

Per-user data always travels in `drivers.MCPCallOptions`; the handle is shared by all users of an instance and keys its caches by the resolved credentials.
