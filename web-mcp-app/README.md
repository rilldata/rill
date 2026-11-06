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
   contract produced by `runtime/ai/create_chart.go`. A result carries `input`: the effective Flint
   `chart_spec` (chart type, channel bindings with the server's top-N `limit`s, properties) and the
   metrics view binding (`data`: metrics view, time range, comparison range, filter, grain) the rows were
   queried with, plus `semantic_types` and `field_display_names`, `data` (rows keyed by field), `fields`
   (display names and formats), `domain_values`, resolved time ranges, `theme`, `open_url`, `truncated`
   and `rows_truncated`. (Rill's own chat gets `chart_type` and `spec` without data instead, and renders
   them itself; that result never reaches this widget.)
3. `src/lib/adapt.ts` turns the payload into what the surface renders from, plus the title, caption and
   data table: `src/lib/flint-input.ts` strips the title, subtitle and `limit` keys, drops channels the
   chart type does not take, fills in the semantic types, and folds a comparison (`<measure>_prev`
   columns) into long-form rows with a `rill_period` series (a bar chart becomes a grouped bar chart so
   the periods sit side by side; areas are layered). The measures get `MetricsViewSpecMeasure`s for
   Rill's value formatters, time series are zero-filled per series over the resolved range, time values
   are converted to the resolved time zone, and the project theme is resolved for the current mode. Flint
   colors series in the order they first appear, so time-ordered rows are put in the server's top-N
   order (`domain_values`) and a refresh keeps every series' color. The caption says "vs. previous
   period" only when the chart draws the fold.
4. `src/components/ChartSurface.svelte` measures its box and compiles the spec with web-common's
   `compileFlintSpec` (the same path canvas custom viz components take: Flint derives scales, axes,
   formats, sorting, stacking and layout from the data and the box; measure axes are routed through
   Rill's formatters), then renders the Vega-Lite with `VegaLiteRenderer` on a canvas using the CSP-safe
   expression interpreter (no `eval`).
5. The inline card (`ChartCard.svelte`) shows the title, caption, chart and two actions: "Open in Rill"
   (`ui/open-link` with the result's `open_url`, http(s) only) and "Fullscreen"
   (`ui/request-display-mode`). `FullscreenView.svelte` adds the Chart/Data tabs, a data table and the
   data controls for time grain, top-N and comparison. Each data control change calls the app-only
   `chart_data` tool through the host (with the patched `chart_spec`, `data` and the caller's
   `semantic_types`; changes in quick succession are debounced and a newer change drops the in-flight
   response). The response is validated and applied only when the chart can show
   it; otherwise, or when the call fails, the previous data stays, the controls roll back and a notice
   says so. The widget then sends `ui/update-model-context` with a one-line description of the current
   view, so the model can answer questions about what the chart shows. The chart also gets
   presentation controls, behind a "Chart options" disclosure so the chart keeps its height, derived
   from Flint's own option model (`src/lib/flint-options.ts`: Flint's chart-type transitions, which
   keep the chart's meaning, such as a line to an area or a grouped bar to a stacked one, and only
   the ones that draw the comparison while one is shown; Flint's arrangements of the same bindings;
   applicable binary and discrete chart properties); those recompile the chart client-side, without a
   tool call, and also update the model context. A chart-type move is applied as a plain chart spec,
   so the comparison fold and a later `chart_data` call see the type that is drawn.

The widget has no network access of its own and holds no credentials: data arrives through the host.

## Layout

```
widget.html           entry (single-file build output: dist/widget.html)
src/App.svelte        host session + state machine
src/lib/host.svelte.ts  reactive wrapper around the ext-apps App
src/lib/contract.ts   wire contract types and validation
src/lib/adapt.ts      payload -> chart inputs (rows, formatters, time zone, zero-fill, theme)
src/lib/flint-input.ts    result -> Flint compile input (stripped spec, semantics, comparison fold)
src/lib/controls.ts   data controls (grain, top-N, comparison): options, chart_data arguments, model text
src/lib/flint-options.ts  presentation controls from Flint's option model (chart type, layout, properties)
src/lib/notices.ts    the banner for values the server or Flint left out
src/lib/theme.ts      host theme and style variables -> Rill tokens and Vega config
src/components/       ChartCard, ChartSurface, FullscreenView, DataTable, SegmentedControl, states
harness/              fake MCP Apps host for development (fixtures, theme and display-mode toggles)
tests/                Playwright suite driving dist/widget.html under a fake host
scripts/              size budget check, fixture generator, svelte-check wrapper
```

## Build, develop, test

- `npm run build -w web-mcp-app`: Vite + `vite-plugin-singlefile` produce `dist/widget.html` and
  `scripts/check-size.mjs` enforces the gzip budget (650 KB; currently about 565 KB, most of it Vega,
  about 60 KB of it Flint's Vega-Lite backend).
  `make runtime.mcpapp.embed` builds and copies the file into `runtime/ai/mcpapp/embed/`.
- `npm run dev -w web-mcp-app` rebuilds on change; run the runtime with `RILL_MCP_APP_DIR=web-mcp-app/dist`
  to serve the latest build without restarting.
- `npm run harness -w web-mcp-app` opens the fake host with every fixture, light and dark mode, locales
  and display modes; no AI product needed.
- `npm run test -w web-mcp-app` (vitest: contract, adapter, controls, theme, and the Flint render path
  per chart fixture, which also checks the fixtures' chart types and channels against the server's
  catalog in `runtime/ai/chartquery/flint_catalog.json` and that the catalog's `flint_version` is the
  installed `flint-chart`) and `npm run test:e2e -w web-mcp-app` (Playwright: rendering of every
  fixture, resizing, theme flips, fullscreen, data and presentation controls, model-context text, error
  and empty states). `npm run check -w web-mcp-app` runs svelte-check on the widget files.
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
