import { createEventBinding } from "@rilldata/web-common/lib/event-emitter.ts";
import { readSSEStream, type SSEMessage } from "./sse-protocol";

export type { SSEMessage };

/**
 * HTTP error thrown by SSEFetchClient when the initial response is non-2xx.
 */
export class SSEHttpError extends Error {
  public readonly status: number;
  public readonly statusText: string;

  constructor(status: number, statusText: string) {
    super(`HTTP ${status}: ${statusText}`);
    this.name = "SSEHttpError";
    this.status = status;
    this.statusText = statusText;

    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, SSEHttpError);
    }
  }
}

type SSEFetchClientEvents = {
  message: SSEMessage;
  error: Error;
  close: void;
  open: void;
};

/**
 * Handles the transport layer for an SSE stream: fetch + AbortController
 * + JWT header + piping the response body through the pure protocol parser.
 *
 * Does not interpret events, does not reconnect. Higher layers (SSEConnection,
 * SSESubscriber) own those concerns.
 */
export class SSEFetchClient {
  private abortController: AbortController | undefined;
  // Monotonic token for start() sessions. When a newer start() happens,
  // async work from older sessions must no-op to avoid stale open/error/close
  // events leaking into the current session.
  private activeRunId = 0;

  private readonly events = createEventBinding<SSEFetchClientEvents>();
  public readonly on = this.events.on;
  public readonly once = this.events.once;

  public async start(
    url: string,
    options: {
      method?: "GET" | "POST";
      body?: Record<string, unknown>;
      headers?: Record<string, string>;
      getJwt?: () => string | undefined;
    } = {},
  ): Promise<void> {
    const runId = ++this.activeRunId;
    const isCurrentRun = () => runId === this.activeRunId;
    // Abort any previous transport immediately. Its async cleanup may still
    // complete later, so lifecycle emits below are gated by isCurrentRun().
    this.stop();

    const {
      method = "GET",
      body,
      headers: customHeaders = {},
      getJwt,
    } = options;

    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...customHeaders,
    };

    try {
      const jwt = getJwt?.();
      if (jwt) {
        headers["Authorization"] = `Bearer ${jwt}`;
      }

      this.abortController = new AbortController();

      const response = await fetch(url, {
        method,
        headers,
        ...(body ? { body: JSON.stringify(body) } : {}),
        signal: this.abortController.signal,
      });

      if (!response.ok) {
        throw new SSEHttpError(response.status, response.statusText);
      }

      if (!response.body) {
        throw new Error("No response body");
      }

      // Another start() replaced this run while fetch was in-flight.
      if (!isCurrentRun()) return;
      this.events.emit("open");

      for await (const message of readSSEStream(response.body)) {
        if (!isCurrentRun()) return;
        this.events.emit("message", message);
      }
    } catch (error) {
      if (!isCurrentRun()) return;
      if (error?.name !== "AbortError") {
        const errorArg =
          error instanceof Error ? error : new Error(String(error));
        this.events.emit("error", errorArg);
      }
    } finally {
      // Only the currently active run should emit close and clear transport.
      if (!isCurrentRun()) return;
      this.stop();
      this.events.emit("close");
    }
  }

  public stop(): void {
    if (this.abortController) {
      this.abortController.abort("SSE stream stopped by client");
      this.abortController = undefined;
    }
  }

  /**
   * Stop streaming and drop all listeners. Call when the client is no longer
   * needed to prevent leaks.
   */
  public cleanup(): void {
    this.stop();
    this.events.clearListeners();
  }

  public isStreaming(): boolean {
    return this.abortController !== undefined;
  }
}
