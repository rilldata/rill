// Minimal MCP Apps host rendered as a standalone page for the Playwright smoke tests.
// Mirrors the protocol subset of harness/host.ts: initialize handshake, tool-input and
// tool-result push, host-context changes, and answers to the widget's requests.
// Every message from the widget is recorded on `window.__messages`.

export interface FakeHostOptions {
  fixture: unknown;
  theme?: "light" | "dark";
  displayMode?: "inline" | "fullscreen";
  locale?: string;
  width?: number;
  styles?: Record<string, string>;
  /** Result returned for `tools/call`; defaults to the fixture's data fields. */
  chartDataResult?: unknown;
}

declare global {
  interface Window {
    __messages: Array<{ method?: string; id?: unknown; params?: unknown }>;
    __host: {
      setTheme(theme: "light" | "dark"): void;
      setDisplayMode(mode: "inline" | "fullscreen"): void;
      setContext(partial: Record<string, unknown>): void;
    };
  }
}

export function fakeHostPage(options: FakeHostOptions): string {
  const serialized = JSON.stringify(options).replace(/</g, "\\u003c");
  return `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Fake MCP Apps host</title>
    <style>
      html, body { margin: 0; padding: 0; }
      body { padding: 12px; background: #fafafa; }
      iframe { display: block; width: 100%; border: 0; height: 420px; }
      iframe.fullscreen { position: fixed; inset: 0; width: 100vw; height: 100vh; }
    </style>
  </head>
  <body>
    <iframe id="widget" title="widget" src="/widget.html" sandbox="allow-scripts"></iframe>
    <script id="options" type="application/json">${serialized}</script>
    <script>
      const options = JSON.parse(document.getElementById("options").textContent);
      const iframe = document.getElementById("widget");
      const state = {
        theme: options.theme || "light",
        displayMode: options.displayMode || "inline",
        locale: options.locale || "en-US",
        width: options.width || 640,
      };
      const DATA_KEYS = ["time_grain", "data", "fields", "semantic_types", "field_display_names",
        "domain_values", "totals",
        "resolved_time_range", "resolved_comparison_time_range", "truncated", "rows_truncated",
        "truncation_warning"];
      window.__messages = [];
      function send(message) {
        iframe.contentWindow.postMessage(message, "*");
      }
      function respond(id, result) {
        send({ jsonrpc: "2.0", id, result });
      }
      function context() {
        return {
          theme: state.theme,
          displayMode: state.displayMode,
          availableDisplayModes: ["inline", "fullscreen"],
          containerDimensions: { width: state.width, maxHeight: 480 },
          locale: state.locale,
          timeZone: "UTC",
          platform: "web",
          safeAreaInsets: { top: 0, right: 0, bottom: 0, left: 0 },
          styles: options.styles ? { variables: options.styles } : undefined,
        };
      }
      function contextChanged(partial) {
        send({ jsonrpc: "2.0", method: "ui/notifications/host-context-changed", params: partial });
      }
      function dataPayload() {
        const source = (options.fixture.result && options.fixture.result.structuredContent) || {};
        const payload = {};
        for (const key of DATA_KEYS) if (key in source) payload[key] = source[key];
        return payload;
      }
      window.__host = {
        setTheme(theme) { state.theme = theme; contextChanged({ theme }); },
        setDisplayMode(mode) {
          state.displayMode = mode;
          iframe.classList.toggle("fullscreen", mode === "fullscreen");
          contextChanged({ displayMode: mode });
        },
        setContext(partial) { contextChanged(partial); },
      };
      window.addEventListener("message", (event) => {
        if (event.source !== iframe.contentWindow) return;
        const message = event.data;
        if (!message || message.jsonrpc !== "2.0") return;
        window.__messages.push(message);
        const id = message.id;
        switch (message.method) {
          case "ui/initialize":
            respond(id, {
              protocolVersion: message.params.protocolVersion,
              hostInfo: { name: "fake-host", version: "0.0.0" },
              hostCapabilities: { openLinks: {}, serverTools: {} },
              hostContext: context(),
            });
            break;
          case "ui/notifications/initialized":
            send({ jsonrpc: "2.0", method: "ui/notifications/tool-input", params: { arguments: options.fixture.arguments } });
            setTimeout(() => {
              send({ jsonrpc: "2.0", method: "ui/notifications/tool-result", params: options.fixture.result });
            }, 20);
            break;
          case "ui/open-link":
            respond(id, {});
            break;
          case "ui/request-display-mode": {
            const mode = message.params.mode;
            respond(id, { mode });
            window.__host.setDisplayMode(mode);
            break;
          }
          case "tools/call":
            respond(id, options.chartDataResult || {
              content: [{ type: "text", text: "ok" }],
              structuredContent: dataPayload(),
            });
            break;
          case "ui/update-model-context":
          case "ui/message":
            respond(id, {});
            break;
          default:
            if (id !== undefined && message.method) {
              send({ jsonrpc: "2.0", id, error: { code: -32601, message: "Method not found" } });
            }
        }
      });
    </script>
  </body>
</html>`;
}
