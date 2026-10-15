# mcp

Connector driver for external MCP (Model Context Protocol) servers. It lets Rill's AI agents discover and call tools on a remote server over streamable HTTP.

## Files

- `mcp.go`: driver registration, config validation, the handle with its client sessions and tool caches (keyed by credentials), and result flattening.
- `transport.go`: the HTTP transport, including the dialer that refuses private addresses in hosted deployments and the round tripper that injects the connector's headers.

## How it fits together

- The parser (`runtime/parser/parse_connector.go`) validates the connector and lifts its `security` block into security rules on the connector spec. Header templates (`{{ .env.* }}`) are resolved by the runtime like any other templated connector property.
- The connector reconciler pings the server, which connects with the connector's headers, so unreachable servers and rejected credentials surface as reconcile errors.
- The AI layer (`runtime/ai/external_tools.go`) discovers servers per session, applies the security policy through the security engine, and compiles the tools for the agents.

Credentials are project-scoped: every user of the server acts with the connector's identity on the server. That is why a connector without a `security` block is only available to admins.
