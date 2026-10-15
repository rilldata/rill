---
title: "External MCP Servers"
description: Give Rill's AI agents tools from external MCP servers
sidebar_label: "External MCP Servers"
sidebar_position: 06
---

Rill's AI agents can call tools on external [MCP (Model Context Protocol)](https://modelcontextprotocol.io) servers, such as an issue tracker, a CRM, or an internal service your team runs. You declare each server as a connector in your project; the analyst and developer agents then discover its tools and use them alongside Rill's built-in tools when a question needs data or actions outside Rill.

This is the outbound counterpart of the [Rill MCP Server](/guide/ai/mcp), which lets external AI clients query Rill.

## Add a server

Create a connector file with `driver: mcp`, the server's streamable HTTP endpoint, and its credentials as headers. Keep secrets in `.env` or in Rill Cloud variables and reference them with `{{ .env.* }}`:

```yaml
# connectors/issues.yaml
type: connector
driver: mcp
url: https://mcp.example.com/mcp
display_name: Issues
ai_instructions: >
  Use it to look up issues and projects. Search before creating anything.
  Never create or update issues unless the user explicitly asks.
headers:
  Authorization: "Bearer {{ .env.ISSUES_MCP_TOKEN }}"
```

Rill connects to the server when the project reconciles, so an unreachable server or rejected credentials show up as an error on the connector. The agents are told about each server's display name and instructions, and its tools appear in the chat as `<display name>: <tool>`.

The credentials are shared: every user of the server acts with the connector's identity on the server. That is why, by default, only admins may use the server. To open it to every user with AI access, add:

```yaml
security:
  access: true
```

Use a token whose permissions you are comfortable granting to everyone the `security` block allows.

## Control who may use a server

The `security` block decides who may use the server, which agents may use it, and which of its tools each user may call. It uses the same syntax as [metrics view security policies](/developers/build/metrics-view/security), with tool names in place of dimension and measure names:

```yaml
security:
  # Who may use the server.
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

Without a `security` block, only admins may use the server. A `security` block without `access` denies everyone, like metrics views. Users who are denied do not see the server at all, and denied tools are never offered to the agents. Policies are evaluated on every tool call, and you can test them in Rill Developer with [mock users](/developers/build/metrics-view/security#test-your-policies).

## Reference

| Property | Description |
| --- | --- |
| `url` | Streamable HTTP endpoint of the server. Must use `https` in Rill Cloud. |
| `display_name` | Name shown in the chat and to the agents. Defaults to the connector name. |
| `ai_instructions` | Guidance for the agents on when and how to use the server. |
| `headers` | Headers sent on every request. Supports `{{ .env.* }}`. |
| `timeout` | Timeout per tool call (default `60s`). |
| `connect_timeout` | Timeout for connecting to the server (default `10s`). |
| `max_result_bytes` | Maximum size of a tool result passed to the AI; larger results are truncated (default `131072`). |
| `security` | Access policy, as described above. |

Tool descriptions and results from external servers are treated as untrusted content: the agents are instructed not to follow instructions found inside them. Each chat message may make at most 25 external tool calls, at most 80 external tools are offered to the agents across all servers, and the tools of external servers are not exposed through the Rill MCP Server.
