# mcpapp: interactive charts for MCP hosts

This package makes Rill's MCP server render charts inside MCP hosts such as Claude (web, desktop, mobile),
ChatGPT, Cursor and VS Code. It implements the server side of the **MCP Apps** extension and owns the widget
bundle, the resource and tool metadata that describe it, and the capability that advertises it.

## Standards

| Layer | Standard | Where Rill implements it |
| - | - | - |
| Transport and protocol | [Model Context Protocol](https://modelcontextprotocol.io/specification), streamable HTTP transport, JSON-RPC 2.0; the server negotiates protocol versions `2025-11-25`, `2025-06-18`, `2025-03-26` and `2024-11-05` | `runtime/server/mcp.go` (runtime endpoint), `admin/server/mcp.go` (unified Rill Cloud endpoint), both on `github.com/modelcontextprotocol/go-sdk` v1.4.1 |
| Tools with structured results | MCP `tools/call` with `content` and `structuredContent` validated against the tool's `outputSchema` | `runtime/ai/ai.go` (`RegisterTool`), `runtime/ai/create_chart.go`, `runtime/ai/chart_data.go` |
| Interactive UI | [MCP Apps](https://modelcontextprotocol.io/docs/extensions/apps), the official MCP extension [SEP-1865](https://github.com/modelcontextprotocol/modelcontextprotocol/pull/1865) (specification revision `2026-01-26`, extension identifier `io.modelcontextprotocol/ui`, source in [modelcontextprotocol/ext-apps](https://github.com/modelcontextprotocol/ext-apps)) | this package, `runtime/ai/mcp.go`, `admin/server/mcp.go`; the widget in `web-mcp-app/` |
| ChatGPT compatibility | OpenAI's plugin UI metadata layered on MCP Apps (`openai/*` `_meta` keys); ChatGPT reads the standard `_meta.ui` fields first | `WidgetResource` in this package; `openai/toolInvocation/*` strings on the tools |

Nothing here is Rill-specific protocol: any host that implements MCP Apps renders the chart, and any host
that does not falls back to the tool's text result.

## How it works

1. **Capability.** When the widget is served, the server's `initialize` result advertises the extension:
   `capabilities.extensions["io.modelcontextprotocol/ui"] = {"mimeTypes": ["text/html;profile=mcp-app"]}`
   (`ServerCapabilities`). Hosts that support MCP Apps advertise the same extension in their request;
   the server does not gate on it because the stateless transport does not carry the client's
   capabilities on later requests, and the extension is harmless to hosts that ignore it.
2. **UI resource.** The widget is one self-contained HTML document, served as an MCP resource with URI
   `ui://rill/widget/<hash>.html` and MIME type `text/html;profile=mcp-app` (`WidgetResource`,
   `Register`). `<hash>` is the first 12 hex characters of the SHA-256 of the bundle: hosts cache
   resources by URI, so a new build gets a new URI and is refetched. A resource template also serves the
   current bundle under any `ui://rill/widget/*.html` URI, because hosts keep the tool list of the build
   they last saw and request that URI after a release.
3. **Tool linkage.** `create_chart` carries `_meta.ui.resourceUri` pointing at the resource
   (`WidgetToolMeta`). When the model calls it, the host reads the resource, renders it in a sandboxed
   iframe (a native WebView on mobile) and delivers the call's arguments and result to the iframe through
   the extension's `ui/notifications/tool-input` and `ui/notifications/tool-result` messages.
   On MCP servers the tool takes the [Flint](https://github.com/microsoft/flint-chart) shape of a chart
   (`ai.MCPSpecProvider`, `runtime/ai/chart_flint.go`): `chart_spec` with a `chartType` from the embedded
   catalog (`runtime/ai/chartquery/flint_catalog.json`, generated from the installed `flint-chart`
   package) and `encodings` binding channels to fields, plus a `data` block naming the metrics view, time
   range, comparison, filter and grain. The in-app chat keeps Rill's own chart spec (`Spec`).
4. **Data passing.** The tool result has two parts. `content` is a markdown rendering (title, a table of
   the first rows, the "Open in Rill" link) for the model and for hosts without MCP Apps.
   `structuredContent` is the full chart payload the widget renders from: the effective input (the chart
   spec with the applied top-N limits and the data block with the resolved grain), rows keyed by field,
   field metadata with formats, the Flint semantic types and display names derived from the metrics view
   (`runtime/ai/chart_semantics.go`), domain values, resolved time ranges, the project theme and the deep
   link. The contract is the JSON form of `CreateChartResult` in `runtime/ai/create_chart.go`. The
   widget compiles the spec with Flint and renders the Vega-Lite it produces; it needs no network access
   and never talks to the runtime directly.
5. **App-only tools.** `chart_data` carries `_meta.ui.visibility: ["app"]` (`AppOnlyToolMeta`). Hosts
   hide it from the model and only let the widget call it, through the host's `tools/call` bridge, when
   the user changes the time grain, the top-N limit or the comparison inside the widget. Those calls are
   not persisted as conversation messages and are not counted as billable tool calls.
6. **Host context.** The host tells the widget its theme (light or dark), CSS variables, display mode
   (`inline` or `fullscreen`), container size, locale and time zone. The widget sends `ui/open-link` for
   "Open in Rill" and `ui/update-model-context` so the model knows what the chart currently shows.

Request flow for one chart:

```
model -> tools/call create_chart ------------> runtime: validate the chart spec against the catalog and the metrics view,
                                               plan the queries from the bound fields (chartquery.BuildFlint), run them
                                               through the metrics resolver, return content (markdown) + structuredContent
host  -> resources/read ui://rill/widget/.. -> runtime (or admin): widget HTML (cached by the host afterwards)
host  -> iframe: tool-input, tool-result ----> widget compiles the spec with Flint and renders the Vega-Lite with Rill's theme
widget -> host -> tools/call chart_data -----> runtime: fresh rows for the changed spec (app-only, ephemeral)
```

## Security

- The host enforces a Content Security Policy built from the resource's `_meta.ui.csp`. Rill declares no
  `connectDomains` and no `frameDomains`; the only external origin is `https://assets.claude.ai` in
  `resourceDomains`, which lets the widget load the host font Claude provides. Everything else is inline.
- The widget runs with the same security claims as the conversation: every query behind `create_chart`
  and `chart_data` goes through the metrics resolver with the caller's claims, so metrics view security
  policies, row filters and field access apply unchanged.
- No tokens are ever placed in the result or the widget; data flows only through the host's MCP bridge.

## Gating and rollout

- The `chat_charts` feature flag (default on) is the single gate: it enables the chart tools and, on the
  runtime's own endpoint, the widget. Clients without MCP Apps get the markdown table and the link.
- The widget is served whenever the binary carries the bundle. The unified `/v1/mcp` endpoint of the admin
  service cannot evaluate per-project flags, so it serves the widget unconditionally and leaves the
  per-project check to the runtime, which rejects `create_chart` calls for projects with charts disabled.
- `chart_data` may only be called by the widget, so both endpoints advertise it exactly when they serve the
  widget (`Session.MCPServer` in runtime/ai/mcp.go and `mcpForwardedToolSpecs` in admin/server/mcp.go);
  `create_chart` is advertised regardless.
- `chart_data` calls count in the `tool_call` metric like every other data tool, since they run real
  queries; they are not persisted in the session, since they are widget plumbing rather than conversation.
- Both servers register the resource and the `_meta.ui` entries with the same helpers from this package
  (`Register`, `WidgetToolMeta`, `AppOnlyToolMeta`, `ToolWithMeta`, `ServerCapabilities`), so the two
  endpoints advertise identical metadata.

## Building and serving the bundle

- `embed/widget.html` is the build output of the `web-mcp-app` workspace and is not committed.
  `make runtime.mcpapp.embed` builds it and copies it into `embed/`; `make cli` runs that through
  `cli.prepare`, and the Rill Cloud image build runs it before `go build` embeds the file.
- `Widget()` returns the embedded bundle, or the file named by `RILL_MCP_APP_DIR` during development
  (re-read whenever it changes, so `npm run dev -w web-mcp-app` gives a hot-reload loop). When no bundle is
  available the server simply does not advertise the widget.

## Testing

- `go test ./runtime/ai/mcpapp/` covers the metadata (structpb-safe types), URI hashing, the resource
  template for stale URIs and the `_meta` merge helper.
- `go test ./runtime/server/ -run MCP` and `go test ./admin/server/ -run 'MCP|CallRuntimeTool'` cover
  capability negotiation, `resources/list` and `resources/read`, tool metadata with the flag on and off,
  the markdown fallback and the pass-through of results on the unified endpoint.
- Protocol check against a running server:
  `npx @modelcontextprotocol/inspector --cli http://localhost:9009/mcp --transport http --method resources/list`.
- End to end in Claude: add the server as a custom connector (through a tunnel for a local server, with the
  Host header rewritten to `localhost:<port>`), or through `mcp-remote` in Claude Desktop; in ChatGPT,
  Plugins > Add custom MCP server.

## References

- MCP Apps specification: https://modelcontextprotocol.io/docs/extensions/apps and
  https://github.com/modelcontextprotocol/ext-apps/blob/main/specification/2026-01-26/apps.mdx
- MCP Apps SDK (used by the widget): https://modelcontextprotocol.github.io/ext-apps/api/
- Claude: https://claude.com/docs/connectors/building/mcp-apps/getting-started and the design guidelines
  under the same path
- ChatGPT: https://developers.openai.com/plugins/build/chatgpt-ui and
  https://developers.openai.com/plugins/reference
- MCP specification: https://modelcontextprotocol.io/specification
