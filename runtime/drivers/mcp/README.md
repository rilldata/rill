# mcp

Connector driver for external MCP (Model Context Protocol) servers. It lets Rill's AI agents discover and call tools on a remote server over streamable HTTP.

## Files

- `mcp.go`: driver registration, config validation, the handle with its per-credential session and tool caches, the OAuth configuration hash, and result flattening.
- `transport.go`: the HTTP transport, including the dialer that refuses private addresses in hosted deployments and the round tripper that injects credentials and records authorization failures.

## How it fits together

- The parser (`runtime/parser/parse_connector.go`) validates the connector, lifts its `security` block into security rules on the connector spec, and leaves header templates unresolved because they may reference the requesting user.
- The connector reconciler pings the server. Which tools a server exposes depends on the user, so the tool catalog is always listed per user by the AI layer.
- The AI layer (`runtime/ai/external_tools.go`) discovers servers per session, applies the security policy through the security engine, compiles the tools for the agents, and resolves per-user OAuth tokens through the admin service.
- The admin service (`admin/server/connector_oauth.go`) runs the browser-based authorization flow with `runtime/pkg/oauthclient` and stores each user's token. Per-user OAuth is therefore only available in Rill Cloud; Rill Developer lists such servers but cannot connect them.

Per-user data always travels in `drivers.MCPCallOptions`; the handle is shared by all users of an instance, holds no OAuth state, and keys its caches by the resolved credentials.
