// Fake MCP Apps host for developing the widget without Claude or ChatGPT.
// It answers `ui/initialize`, pushes a fixture as `tool-input` + `tool-result`, relays
// host-context changes (theme, locale, display mode, size), answers `ui/open-link`,
// `ui/request-display-mode`, `ui/update-model-context` and `tools/call` (with the
// fixture's own data), and logs every message in both directions.

interface Fixture {
  name: string;
  description?: string;
  expect?: string;
  arguments: Record<string, unknown>;
  result: {
    isError?: boolean;
    content?: unknown[];
    structuredContent?: Record<string, unknown>;
  };
}

interface JsonRpcMessage {
  jsonrpc: "2.0";
  id?: number | string;
  method?: string;
  params?: Record<string, unknown>;
  result?: unknown;
  error?: { code: number; message: string };
}

const HOST_STYLE_VARIABLES = {
  light: {
    "--color-background-primary": "#ffffff",
    "--color-background-secondary": "#f4f4f5",
    "--color-background-tertiary": "#e8e8ea",
    "--color-text-primary": "#1f1f1f",
    "--color-text-secondary": "#4b4b4f",
    "--color-text-tertiary": "#71717a",
    "--color-text-inverse": "#ffffff",
    "--color-border-primary": "#e4e4e7",
    "--color-border-secondary": "#d4d4d8",
    "--border-radius-md": "10px",
    "--font-sans": 'ui-sans-serif, system-ui, "Segoe UI", Roboto, sans-serif',
  },
  dark: {
    "--color-background-primary": "#1f1f1f",
    "--color-background-secondary": "#2a2a2a",
    "--color-background-tertiary": "#343434",
    "--color-text-primary": "#f2f2f2",
    "--color-text-secondary": "#c9c9c9",
    "--color-text-tertiary": "#9a9a9a",
    "--color-text-inverse": "#111111",
    "--color-border-primary": "#3a3a3a",
    "--color-border-secondary": "#474747",
    "--border-radius-md": "10px",
    "--font-sans": 'ui-sans-serif, system-ui, "Segoe UI", Roboto, sans-serif',
  },
};

const DATA_KEYS = [
  "time_grain",
  "data",
  "fields",
  "semantic_types",
  "field_display_names",
  "domain_values",
  "totals",
  "resolved_time_range",
  "resolved_comparison_time_range",
  "truncated",
  "rows_truncated",
  "truncation_warning",
];

const fixtures = import.meta.glob<Fixture>("./fixtures/*.json", {
  eager: true,
  import: "default",
});

const $ = <T extends HTMLElement>(id: string) =>
  document.getElementById(id) as T;
const iframe = $<HTMLIFrameElement>("widget");
const frame = $<HTMLDivElement>("frame");
const log = $<HTMLPreElement>("log");
const fixtureSelect = $<HTMLSelectElement>("fixture");
const themeSelect = $<HTMLSelectElement>("theme");
const localeSelect = $<HTMLSelectElement>("locale");
const displaySelect = $<HTMLSelectElement>("display");
const widthInput = $<HTMLInputElement>("width");
const widthValue = $<HTMLSpanElement>("width-value");
const hostStyles = $<HTMLInputElement>("host-styles");

for (const path of Object.keys(fixtures).sort()) {
  const option = document.createElement("option");
  option.value = path;
  option.textContent = path.replace("./fixtures/", "").replace(".json", "");
  fixtureSelect.append(option);
}

function currentFixture(): Fixture {
  return fixtures[fixtureSelect.value];
}

function append(direction: "in" | "out", message: unknown) {
  const line = document.createElement("div");
  line.className = direction;
  const text = JSON.stringify(message);
  line.textContent = `${direction === "out" ? "host →" : "app  →"} ${
    text.length > 400 ? `${text.slice(0, 400)}…` : text
  }`;
  log.append(line);
  log.scrollTop = log.scrollHeight;
}

function send(message: JsonRpcMessage) {
  append("out", message);
  iframe.contentWindow?.postMessage(message, "*");
}

function respond(id: number | string, result: unknown) {
  send({ jsonrpc: "2.0", id, result });
}

function hostContext() {
  const theme = themeSelect.value as "light" | "dark";
  return {
    theme,
    displayMode: displaySelect.value,
    availableDisplayModes: ["inline", "fullscreen"],
    containerDimensions: { width: Number(widthInput.value), maxHeight: 520 },
    locale: localeSelect.value,
    timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    platform: "web",
    safeAreaInsets: { top: 0, right: 0, bottom: 0, left: 0 },
    ...(hostStyles.checked && {
      styles: { variables: HOST_STYLE_VARIABLES[theme] },
    }),
  };
}

function sendContextChanged(partial: Record<string, unknown>) {
  send({
    jsonrpc: "2.0",
    method: "ui/notifications/host-context-changed",
    params: partial,
  });
}

function dataPayload(fixture: Fixture) {
  const source = fixture.result.structuredContent ?? {};
  return Object.fromEntries(
    DATA_KEYS.filter((key) => key in source).map((key) => [key, source[key]]),
  );
}

window.addEventListener("message", (event) => {
  if (event.source !== iframe.contentWindow) return;
  const message = event.data as JsonRpcMessage;
  if (!message || message.jsonrpc !== "2.0") return;
  append("in", message);

  if (message.method === "ui/notifications/size-changed") {
    const height = (message.params as { height?: number })?.height;
    if (height && displaySelect.value !== "fullscreen") {
      iframe.style.height = `${height}px`;
    }
    return;
  }

  const id = message.id;
  switch (message.method) {
    case "ui/initialize":
      if (id === undefined) return;
      respond(id, {
        protocolVersion: (message.params as { protocolVersion: string })
          .protocolVersion,
        hostInfo: { name: "rill-harness", version: "0.0.0" },
        hostCapabilities: { openLinks: {}, serverTools: {}, logging: {} },
        hostContext: hostContext(),
      });
      return;
    case "ui/notifications/initialized": {
      const fixture = currentFixture();
      send({
        jsonrpc: "2.0",
        method: "ui/notifications/tool-input",
        params: { arguments: fixture.arguments },
      });
      setTimeout(() => {
        send({
          jsonrpc: "2.0",
          method: "ui/notifications/tool-result",
          params: fixture.result,
        });
      }, 300);
      return;
    }
    case "ui/open-link":
      if (id !== undefined) respond(id, {});
      return;
    case "ui/request-display-mode": {
      const mode = (message.params as { mode: string }).mode;
      displaySelect.value = mode;
      applyDisplayMode();
      if (id !== undefined) respond(id, { mode });
      sendContextChanged({ displayMode: mode });
      return;
    }
    case "tools/call": {
      // Echo the fixture's data after a short delay so the loading state is visible.
      setTimeout(() => {
        if (id !== undefined) {
          respond(id, {
            content: [{ type: "text", text: "ok" }],
            structuredContent: dataPayload(currentFixture()),
          });
        }
      }, 700);
      return;
    }
    case "ui/update-model-context":
    case "ui/message":
      if (id !== undefined) respond(id, {});
      return;
    default:
      if (id !== undefined && message.method) {
        send({
          jsonrpc: "2.0",
          id,
          error: {
            code: -32601,
            message: `Method not found: ${message.method}`,
          },
        });
      }
  }
});

function applyDisplayMode() {
  const fullscreen = displaySelect.value === "fullscreen";
  frame.classList.toggle("fullscreen", fullscreen);
  iframe.style.height = fullscreen ? "100vh" : iframe.style.height;
}

function remount() {
  log.textContent = "";
  iframe.style.height = "120px";
  applyDisplayMode();
  iframe.src = `/widget.html?fixture=${encodeURIComponent(fixtureSelect.value)}&t=${Date.now()}`;
}

fixtureSelect.addEventListener("change", remount);
$("reload").addEventListener("click", remount);
themeSelect.addEventListener("change", () => {
  document.body.classList.toggle("dark", themeSelect.value === "dark");
  sendContextChanged({
    theme: themeSelect.value,
    ...(hostStyles.checked && {
      styles: {
        variables: HOST_STYLE_VARIABLES[themeSelect.value as "light" | "dark"],
      },
    }),
  });
});
localeSelect.addEventListener("change", () =>
  sendContextChanged({ locale: localeSelect.value }),
);
displaySelect.addEventListener("change", () => {
  applyDisplayMode();
  sendContextChanged({ displayMode: displaySelect.value });
});
widthInput.addEventListener("input", () => {
  widthValue.textContent = widthInput.value;
  frame.style.width = `${widthInput.value}px`;
  sendContextChanged({
    containerDimensions: { width: Number(widthInput.value), maxHeight: 520 },
  });
});

remount();
