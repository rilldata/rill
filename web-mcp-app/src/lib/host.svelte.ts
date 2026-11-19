import {
  App,
  type McpUiDisplayMode,
  type McpUiHostCapabilities,
  type McpUiHostContext,
} from "@modelcontextprotocol/ext-apps";
import type { ToolResult } from "./contract";

export type SessionStatus =
  | "connecting"
  | "awaiting-input"
  | "awaiting-result"
  | "received"
  | "cancelled"
  | "failed";

export const MODEL_CONTEXT_DEBOUNCE_MS = 500;
const APP_INFO = { name: "rill-chart", version: "1.0.0" };

/**
 * Reactive wrapper around the ext-apps `App` session.
 * Handlers are registered in the constructor, before `connect()`, because hosts push
 * `tool-input` and `tool-result` right after the initialize handshake.
 * Every capability is feature-detected from the handshake; nothing branches on the host name.
 */
export class HostSession {
  status = $state<SessionStatus>("connecting");
  // Host payloads are replaced wholesale and must stay plain objects: a deep `$state` proxy
  // would leak into the Vega-Lite spec (e.g. a multi-measure `fields` array), and Vega
  // structured-clones its specs, which rejects proxies.
  hostContext = $state.raw<McpUiHostContext | undefined>(undefined);
  hostCapabilities = $state.raw<McpUiHostCapabilities | undefined>(undefined);
  toolInput = $state.raw<Record<string, unknown> | undefined>(undefined);
  toolResult = $state.raw<ToolResult | undefined>(undefined);
  error = $state<string | undefined>(undefined);

  readonly app: App;
  private modelContextTimer: ReturnType<typeof setTimeout> | undefined;

  constructor(
    app: App = new App(
      APP_INFO,
      { availableDisplayModes: ["inline", "fullscreen"] },
      { autoResize: true },
    ),
  ) {
    this.app = app;
    app.ontoolinput = (params) => {
      this.toolInput = params.arguments;
      if (this.status === "awaiting-input") this.status = "awaiting-result";
    };
    app.ontoolresult = (result) => {
      this.toolResult = result;
      this.status = "received";
    };
    app.ontoolcancelled = (params) => {
      this.error = params.reason;
      this.status = "cancelled";
    };
    app.onhostcontextchanged = (params) => {
      // The host sends only the fields that changed.
      this.hostContext = { ...(this.hostContext ?? {}), ...params };
    };
    app.onerror = (error) => {
      console.error("rill chart widget: host session error", error);
      this.error = error.message;
      if (this.status === "connecting") this.status = "failed";
    };
  }

  async connect(): Promise<void> {
    try {
      await this.app.connect();
      this.hostContext = this.app.getHostContext();
      this.hostCapabilities = this.app.getHostCapabilities();
      if (this.status === "connecting") this.status = "awaiting-input";
    } catch (error) {
      console.error("rill chart widget: could not connect to the host", error);
      this.error = error instanceof Error ? error.message : String(error);
      this.status = "failed";
    }
  }

  get displayMode(): McpUiDisplayMode {
    return this.hostContext?.displayMode ?? "inline";
  }

  get isFullscreen(): boolean {
    return this.displayMode === "fullscreen";
  }

  get canFullscreen(): boolean {
    return (this.hostContext?.availableDisplayModes ?? []).includes(
      "fullscreen",
    );
  }

  get canOpenLinks(): boolean {
    return this.hostCapabilities?.openLinks !== undefined;
  }

  get canCallServerTools(): boolean {
    return this.hostCapabilities?.serverTools !== undefined;
  }

  async openLink(url: string): Promise<void> {
    try {
      await this.app.openLink({ url });
    } catch (error) {
      console.warn("rill chart widget: open link rejected", error);
    }
  }

  async requestDisplayMode(
    mode: McpUiDisplayMode,
  ): Promise<McpUiDisplayMode | undefined> {
    try {
      const result = await this.app.requestDisplayMode({ mode });
      // Hosts normally also send host-context-changed; applying the granted mode here keeps
      // the UI consistent with hosts that only answer the request.
      this.hostContext = {
        ...(this.hostContext ?? {}),
        displayMode: result.mode,
      };
      return result.mode;
    } catch (error) {
      console.warn("rill chart widget: display mode request rejected", error);
      return undefined;
    }
  }

  callServerTool(
    name: string,
    args: Record<string, unknown>,
  ): Promise<ToolResult> {
    return this.app.callServerTool({ name, arguments: args });
  }

  /** Debounced so a burst of control changes produces one model update. */
  updateModelContext(text: string): void {
    if (this.modelContextTimer) clearTimeout(this.modelContextTimer);
    this.modelContextTimer = setTimeout(() => {
      this.modelContextTimer = undefined;
      this.app
        .updateModelContext({ content: [{ type: "text", text }] })
        .catch((error: unknown) => {
          console.warn(
            "rill chart widget: model context update rejected",
            error,
          );
        });
    }, MODEL_CONTEXT_DEBOUNCE_MS);
  }

  destroy(): void {
    if (this.modelContextTimer) clearTimeout(this.modelContextTimer);
  }
}
