# MCP Apps widget bundle

This directory holds the single-file HTML bundle of the MCP Apps widget that renders Rill charts inside
MCP hosts such as Claude and ChatGPT. It is served as the `ui://rill/widget/<hash>.html` resource of
Rill's MCP server (see `runtime/ai/mcpapp`).

`widget.html` is a build artifact and is not committed:

- `make runtime.mcpapp.embed` builds the `web-mcp-app` workspace and copies `web-mcp-app/dist/widget.html`
  here. `make cli` runs it through the `cli.prepare` target, and the Rill Cloud image build
  (`.github/workflows/rill-cloud.yml`) runs it before the Go build embeds the file.
- When the file is absent (for example in a plain `go build` or `go test`), `mcpapp.Widget()` reports that
  no widget is available and the MCP server does not advertise the resource.
- For local development, set `RILL_MCP_APP_DIR` to a directory containing `widget.html`
  (for example `web-mcp-app/dist` while `npm run dev -w web-mcp-app` is running); the runtime re-reads the
  file whenever it changes.
