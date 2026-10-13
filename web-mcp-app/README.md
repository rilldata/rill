# web-mcp-app: the Rill chart widget for MCP hosts

A single-file HTML app that renders Rill charts inside MCP hosts (Claude web, desktop and mobile, ChatGPT,
Cursor, VS Code) as an **MCP App**. Rill's MCP server serves the built file as the resource
`ui://rill/widget/<hash>.html`; see `runtime/ai/mcpapp/README.md` for the server side.

## Standards

- **MCP Apps** ([specification](https://modelcontextprotocol.io/docs/extensions/apps), extension
  `io.modelcontextprotocol/ui`, revision `2026-01-26`): the widget is the "view" of the extension. It is
  loaded into a sandboxed iframe by the host and talks to it with JSON-RPC over `postMessage`, using the
  official SDK `@modelcontextprotocol/ext-apps` (`App` class). Lifecycle: `ui/initialize` (host context,
  capabilities), `ui/notifications/tool-input`, `ui/notifications/tool-result`,
  `ui/notifications/host-context-changed`, `ui/notifications/size-changed`; actions: `tools/call` (through
  the host), `ui/open-link`, `ui/request-display-mode`, `ui/update-model-context`.
- **Host theming**: the standard host style variables (`--color-background-*`, `--color-text-*`,
  `--color-border-*`, `--font-sans`, radii) with `light-dark()` fallbacks, applied with the SDK's
  `applyDocumentTheme`, `applyHostStyleVariables` and `applyHostFonts`.
- **ChatGPT**: no ChatGPT-specific code; the host's `window.openai` extensions are not used. ChatGPT
  implements the same MCP Apps bridge.

Everything is feature-detected from the host handshake (`availableDisplayModes`, `openLinks`,
`serverTools`); nothing branches on the host's name.

## How it works

1. `src/App.svelte` creates the SDK `App`, registers the handlers (before `connect()`, since hosts push
   the tool input and result right after the handshake) and tracks the session state: connecting,
   waiting for input, waiting for the result, rendered, empty, error.
2. `src/lib/contract.ts` validates the untrusted tool result (`structuredContent`) against the wire
   contract produced by `runtime/ai/create_chart.go`: `widget: "chart"`, `chart_type`, `spec` (with the
   effective limits filled in), `data` (rows keyed by field), `fields` (display names and formats),
   `domain_values`, `totals`, resolved time ranges, `theme`, `open_url`, `truncated`, `rows_truncated`.
3. `src/lib/adapt.ts` turns the payload into what web-common's chart components expect: a
   `ChartDataResult` store, `MetricsViewSpecMeasure`s for Rill's value formatters, time values converted
   to the resolved time zone, zero-filled time buckets, comparison detection (`<measure>_prev` columns),
   domain values and the project theme resolved for the current mode.
4. `src/components/ChartSurface.svelte` mounts `web-common/src/features/components/charts/Chart.svelte`,
   the same component Rill's in-app AI chat and canvas dashboards use. It generates a Vega-Lite spec from
   the Rill chart config and renders it with vega-embed on a canvas, using the CSP-safe expression
   interpreter (no `eval`).
5. The inline card (`ChartCard.svelte`) shows the title, caption, chart and two actions: "Open in Rill"
   (`ui/open-link` with the result's `open_url`) and "Fullscreen" (`ui/request-display-mode`).
   `FullscreenView.svelte` adds the Chart/Data tabs, a data table and the controls for time grain, top-N
   and comparison. Each control change calls the app-only `chart_data` tool through the host and sends
   `ui/update-model-context` with a one-line description of the current view, so the model can answer
   questions about what the chart shows.

The widget has no network access of its own and holds no credentials: data arrives through the host.

## Layout

```
widget.html           entry (single-file build output: dist/widget.html)
src/App.svelte        host session + state machine
src/lib/host.svelte.ts  reactive wrapper around the ext-apps App
src/lib/contract.ts   wire contract types and validation
src/lib/adapt.ts      payload -> chart inputs (rows, formatters, time zone, zero-fill, theme)
src/lib/controls.ts   grain / top-N / comparison options, spec patching, model-context text
src/lib/theme.ts      host theme and style variables -> Rill tokens and Vega config
src/components/       ChartCard, ChartSurface, FullscreenView, DataTable, SegmentedControl, states
harness/              fake MCP Apps host for development (fixtures, theme and display-mode toggles)
tests/                Playwright suite driving dist/widget.html under a fake host
scripts/              size budget check, fixture generator, svelte-check wrapper
```

## Build, develop, test

- `npm run build -w web-mcp-app`: Vite + `vite-plugin-singlefile` produce `dist/widget.html` and
  `scripts/check-size.mjs` enforces the gzip budget (600 KB; currently about 513 KB, most of it Vega).
  `make runtime.mcpapp.embed` builds and copies the file into `runtime/ai/mcpapp/embed/`.
- `npm run dev -w web-mcp-app` rebuilds on change; run the runtime with `RILL_MCP_APP_DIR=web-mcp-app/dist`
  to serve the latest build without restarting.
- `npm run harness -w web-mcp-app` opens the fake host with every fixture, light and dark mode, locales
  and display modes; no AI product needed.
- `npm run test -w web-mcp-app` (vitest: contract, adapter, controls, theme, spec generation per chart
  type) and `npm run test:e2e -w web-mcp-app` (Playwright: rendering of every fixture including results
  captured from a real server, resizing, theme flips, fullscreen, controls, model-context text, error and
  empty states). `npm run check -w web-mcp-app` runs svelte-check on the widget files.
- Run npm commands with Node 20 (the repo's `.nvmrc`); `npm run build:i18n` at the repo root must have run
  once, since the widget uses the shared paraglide messages (`mcp_chart_*` keys in `en.json` and
  `es.json`).

## Host notes

- Claude: custom connectors render MCP Apps on every plan; the first display asks for permission; tool
  results above roughly 150,000 characters are written to a file by Claude and never reach the widget,
  which is why the server keeps payloads small (row and byte budgets with `truncated` flags).
- ChatGPT: the inline widget renders above the model's answer; system fonts and host colors are used.
- Hosts cache the resource by URI; the server serves the current bundle under stale URIs so a redeploy
  does not break open conversations.
