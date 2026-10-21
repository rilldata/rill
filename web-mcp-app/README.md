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
   contract produced by `runtime/ai/create_chart.go`. A **Flint result** carries `input`: the effective
   Flint `chart_spec` (chart type, channel bindings with the server's top-N `limit`s, properties) and the
   metrics view binding (`data`: metrics view, time range, comparison range, filter, grain) the rows were
   queried with, plus `semantic_types` and `field_display_names`. A **Rill result** carries `spec`, Rill's
   own chart config for the 11 built-in types, which older runtimes answer with. Both carry `data` (rows
   keyed by field), `fields` (display names and formats), `domain_values`, resolved time ranges, `theme`,
   `open_url`, `truncated` and `rows_truncated`.
3. `src/lib/adapt.ts` turns either payload into what the surface renders from, plus the title, caption
   and data table: for a Flint result, `src/lib/flint-input.ts` strips the title, subtitle and `limit`
   keys, drops channels the chart type does not take, fills in the semantic types, and folds a comparison
   (`<measure>_prev` columns) into long-form rows with a `rill_period` series (a bar chart becomes a
   grouped bar chart so the periods sit side by side; areas are layered). For a Rill result it builds the
   `ChartDataResult` store web-common's `Chart.svelte` expects. Both get `MetricsViewSpecMeasure`s for
   Rill's value formatters, time values converted to the resolved time zone, and the project theme
   resolved for the current mode.
4. `src/components/ChartSurface.svelte` dispatches: a Flint result goes to `FlintSurface.svelte`, which
   measures its box and compiles the spec with web-common's `compileFlintSpec` (the same path canvas
   custom viz components take: Flint derives scales, axes, formats, sorting, stacking and layout from the
   data and the box; measure axes are routed through Rill's formatters), then renders the Vega-Lite with
   `VegaLiteRenderer` on a canvas using the CSP-safe expression interpreter (no `eval`). A Rill result
   mounts `web-common/src/features/components/charts/Chart.svelte`, the component Rill's in-app AI chat
   and canvas dashboards use.
5. The inline card (`ChartCard.svelte`) shows the title, caption, chart and two actions: "Open in Rill"
   (`ui/open-link` with the result's `open_url`) and "Fullscreen" (`ui/request-display-mode`).
   `FullscreenView.svelte` adds the Chart/Data tabs, a data table and the data controls for time grain,
   top-N and comparison. Each data control change calls the app-only `chart_data` tool through the host
   (with the patched `chart_spec` and `data` on the Flint path; changes in quick succession are debounced
   and a newer change drops the in-flight response) and sends `ui/update-model-context` with
   a one-line description of the current view, so the model can answer questions about what the chart
   shows. Flint charts get a second row of presentation controls derived from Flint's own option model
   (`src/lib/flint-options.ts`: chart types the bindings can move to, the alternative views of the pivot
   surface, applicable binary and discrete chart properties); those recompile the chart client-side,
   without a tool call, and also update the model context.

The widget has no network access of its own and holds no credentials: data arrives through the host.

## Layout

```
widget.html           entry (single-file build output: dist/widget.html)
src/App.svelte        host session + state machine
src/lib/host.svelte.ts  reactive wrapper around the ext-apps App
src/lib/contract.ts   wire contract types and validation
src/lib/adapt.ts      payload -> chart inputs (rows, formatters, time zone, zero-fill, theme)
src/lib/flint-input.ts    Flint result -> compile input (stripped spec, semantics, comparison fold)
src/lib/flint-controls.ts data controls on the Flint path: options, chart_data arguments, model text
src/lib/flint-options.ts  presentation controls from Flint's option model (chart type, views, properties)
src/lib/controls.ts   grain / top-N / comparison options, spec patching, model-context text (Rill results)
src/lib/theme.ts      host theme and style variables -> Rill tokens and Vega config
src/components/       ChartCard, ChartSurface, FlintSurface, FullscreenView, DataTable, SegmentedControl, states
harness/              fake MCP Apps host for development (fixtures, theme and display-mode toggles)
tests/                Playwright suite driving dist/widget.html under a fake host
scripts/              size budget check, fixture generator, svelte-check wrapper
```

## Build, develop, test

- `npm run build -w web-mcp-app`: Vite + `vite-plugin-singlefile` produce `dist/widget.html` and
  `scripts/check-size.mjs` enforces the gzip budget (650 KB; currently about 575 KB, most of it Vega,
  about 60 KB of it Flint's Vega-Lite backend).
  `make runtime.mcpapp.embed` builds and copies the file into `runtime/ai/mcpapp/embed/`.
- `npm run dev -w web-mcp-app` rebuilds on change; run the runtime with `RILL_MCP_APP_DIR=web-mcp-app/dist`
  to serve the latest build without restarting.
- `npm run harness -w web-mcp-app` opens the fake host with every fixture, light and dark mode, locales
  and display modes; no AI product needed.
- `npm run test -w web-mcp-app` (vitest: contract, adapters, controls, theme, spec generation per Rill
  chart type, and the Flint render path per `flint_*` fixture, which also checks the fixtures' chart
  types and channels against the server's catalog in `runtime/ai/chartquery/flint_catalog.json`) and
  `npm run test:e2e -w web-mcp-app` (Playwright: rendering of every fixture including results captured
  from a real server, resizing, theme flips, fullscreen, data and presentation controls, model-context
  text, error and empty states). `npm run check -w web-mcp-app` runs svelte-check on the widget files.
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
