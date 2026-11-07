---
title: "External MCP Servers"
description: Give Rill's AI agents tools from external MCP servers, with per-user or project-scoped credentials
sidebar_label: "External MCP Servers"
sidebar_position: 06
---

Rill's AI agents can call tools on external [MCP (Model Context Protocol)](https://modelcontextprotocol.io) servers, such as Linear, GitHub, Slack, or an internal service your team runs. You declare each server as a connector in your project; the analyst and developer agents then discover its tools and use them alongside Rill's built-in tools when a question needs data or actions outside Rill.

This is the outbound counterpart of the [Rill MCP Server](/guide/ai/mcp), which lets external AI clients query Rill.

## Add a server

Create a connector file with `driver: mcp` and the server's streamable HTTP endpoint:

```yaml
# connectors/linear.yaml
type: connector
driver: mcp
url: https://mcp.linear.app/mcp
display_name: Linear
ai_instructions: >
  Use Linear to look up issues and projects. Search before creating anything.
  Never create or update issues unless the user explicitly asks.
auth: oauth
```

The connector appears in Rill Developer with a side panel that shows the server's status and the tools it exposes to you. In the chat, the plug icon (**External tools**) lists the project's external servers, and the agents are told about each server's display name and instructions.

## Choose how to authenticate

Each server uses one of two credential scopes. The choice is visible in the YAML and reviewable in git.

### Per-user OAuth

Set `auth: oauth` for servers that identify individual users, such as Linear, GitHub, Slack, Notion, or Atlassian. Each user authorizes the server from the chat's **External tools** popover: **Connect** opens the server's authorization page in a popup. If the assistant tries to use a server the user has not connected yet, it stops and shows a Connect card instead of failing; once connected, send your message again.

Per-user OAuth is available in Rill Cloud only. Rill Developer lists such servers but cannot connect them, so test them after deploying. Servers with project-scoped credentials (below) work in both.

Rill discovers the server's authorization server automatically and registers itself as an OAuth client when the server supports dynamic client registration. Servers that require a pre-registered client, such as Slack, take the client credentials from the connector:

```yaml
type: connector
driver: mcp
url: https://mcp.slack.com/mcp
display_name: Slack
auth: oauth
oauth_client_id: "{{ .env.SLACK_MCP_CLIENT_ID }}"
oauth_client_secret: "{{ .env.SLACK_MCP_CLIENT_SECRET }}"
oauth_scopes: [search:read, channels:history]
```

Register `https://admin.rilldata.com/connectors/oauth/callback` as the redirect URL when the server asks for one.

Some authorization servers need extra parameters in the authorization request. Google, for example, issues a refresh token only when asked for offline access, and only on a user's first consent unless it is asked to show the consent screen again. Without a refresh token, users have to connect again each time their access token expires (typically after one hour for Google). Pass such parameters with `oauth_authorization_params`:

```yaml
type: connector
driver: mcp
url: https://drivemcp.googleapis.com/mcp/v1
display_name: Google Drive
auth: oauth
oauth_client_id: "{{ .env.GOOGLE_OAUTH_CLIENT_ID }}"
oauth_client_secret: "{{ .env.GOOGLE_OAUTH_CLIENT_SECRET }}"
oauth_scopes: [https://www.googleapis.com/auth/drive.readonly]
oauth_authorization_params:
  access_type: offline
  prompt: consent
```

With `prompt: consent`, Google shows its consent screen every time a user connects, which only happens again when the user disconnects or the authorization is revoked. Parameters that Rill sets itself or depends on, such as `state`, `scope`, `code_challenge`, `request` or `response_mode`, cannot be set. Changing `oauth_authorization_params` requires users to connect again.

Tokens are never shared between users. Rill Cloud stores each user's token encrypted in the control plane and refreshes it as needed; runtimes only receive short-lived access tokens. When the authorization server advertises the `offline_access` scope, as Microsoft Entra ID does, Rill requests it so that the server issues a refresh token. Rill only adds it to the scopes it discovers: `oauth_scopes` are requested exactly as listed, so include `offline_access` there if the server needs it. Users can disconnect a server at any time from the chat, even while the server is unreachable, which deletes the token.

### Project-scoped credentials

For internal servers or servers that accept a service token, send the credentials as headers. Keep secrets in `.env` or in Rill Cloud variables and reference them with `{{ .env.* }}`:

```yaml
type: connector
driver: mcp
url: https://mcp.internal.example.com/mcp
display_name: Company data
headers:
  Authorization: "Bearer {{ .env.INTERNAL_MCP_TOKEN }}"
  X-Rill-User: "{{ .user.email }}"
```

Header values may also reference attributes of the requesting user with `{{ .user.* }}`, so a server that trusts Rill can enforce its own per-user permissions. A server whose headers reference the user is only available to identified users.

`auth: oauth` and an `Authorization` header cannot be combined.

## Control who may use a server

The `security` block decides who may use the server, which agents may use it, and which of its tools each user may call. It uses the same syntax as [metrics view security policies](/developers/build/metrics-view/security), with tool names in place of dimension and measure names:

```yaml
security:
  # Who may use the server. Omit the whole security block to allow everyone with AI access.
  access: "{{ .user.admin }} OR '{{ .user.domain }}' == 'example.com'"
  # Which agents may use it. Defaults to all agents.
  agents: [analyst]
  # Which tools each user may call. If any include rule is present, unlisted tools are denied.
  include:
    - if: true
      names: [search_issues, get_issue]
    - if: "{{ .user.admin }}"
      names: [create_issue, update_issue]
  # Tools to deny. A matching exclude always wins.
  exclude:
    - if: '{{ not (has "engineering" .user.groups) }}'
      names: [delete_issue]
```

The defaults match metrics views: no `security` block allows every user with AI access, while a `security` block without `access` denies everyone. Users who are denied do not see the server at all, and denied tools are never offered to the agents. Policies are evaluated on every tool call, and you can test them in Rill Developer with [mock users](/developers/build/metrics-view/security#test-your-policies).

## Reference

| Property | Description |
| --- | --- |
| `url` | Streamable HTTP endpoint of the server. Must use `https` in Rill Cloud. |
| `display_name` | Name shown in the chat and to the agents. Defaults to the connector name. |
| `ai_instructions` | Guidance for the agents on when and how to use the server. |
| `auth` | `oauth` for per-user authorization. |
| `oauth_client_id`, `oauth_client_secret`, `oauth_scopes` | Pre-registered client and scopes for servers without dynamic client registration. |
| `oauth_authorization_params` | Extra query parameters for the authorization request, such as Google's `access_type: offline`. |
| `headers` | Headers sent on every request. Supports `{{ .env.* }}` and `{{ .user.* }}`. |
| `timeout` | Timeout per tool call (default `60s`). |
| `connect_timeout` | Timeout for connecting and listing tools (default `10s`). |
| `max_result_bytes` | Maximum size of a tool result passed to the AI; larger results are truncated (default `131072`). |
| `security` | Access policy, as described above. |

Tool descriptions and results from external servers are treated as untrusted content: the agents are instructed not to follow instructions found inside them. Each chat message may make at most 25 external tool calls. The agents load external tools on demand, by name or keyword search, so servers with many tools do not crowd every request; at most 100 external tools can be loaded per conversation. The tools of external servers are not exposed through the Rill MCP Server.
