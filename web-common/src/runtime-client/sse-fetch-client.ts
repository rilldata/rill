import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";

/**
 * Represents a Server-Sent Event message
 */
export interface SSEMessage {
  /** Event type (undefined means default 'message' type) */
  type?: string;
  /** Raw event data */
  data: string;
}

/**
 * HTTP error thrown by SSEFetchClient
 */
export class SSEHttpError extends Error {
  public readonly status: number;
  public readonly statusText: string;

  constructor(status: number, statusText: string) {
    super(`HTTP ${status}: ${statusText}`);
    this.name = "SSEHttpError";
    this.status = status;
    this.statusText = statusText;

    // Maintains proper stack trace for where error was thrown (only available on V8)
    if (Error.captureStackTrace) {
      Error.captureStackTrace(this, SSEHttpError);
    }
  }
}

// ===== SSE PROTOCOL PARSING =====
// These functions handle SSE format parsing and could be extracted
// to a separate module if needed for reuse.

/**
 * Parse a single SSE line and update the accumulating event
 * Mutates the event object for efficiency during streaming
 */
function parseSSELine(line: string, event: Partial<SSEMessage>): void {
  const trimmedLine = line.trim();

  // Skip empty lines (handled separately as event boundaries)
  if (!trimmedLine) return;

  // Skip comments
  if (trimmedLine.startsWith(":")) return;

  // Parse event type
  if (trimmedLine.startsWith("event:")) {
    event.type = trimmedLine.slice(6).trim();
    return;
  }

  // Parse data (can span multiple lines)
  if (trimmedLine.startsWith("data:")) {
    const data = trimmedLine.slice(5).trim();
    event.data = event.data ? event.data + "\n" + data : data;
    return;
  }

  // Note: Could extend to handle "id:" and "retry:" fields if needed
}

/**
 * Check if a line represents an event boundary (empty line)
 */
function isEventComplete(line: string): boolean {
  return line.trim() === "";
}

/**
 * Check if an event has the minimum required data to be emitted
 */
function isValidEvent(event: Partial<SSEMessage>): event is SSEMessage {
  return event.data !== undefined && event.data !== "";
}

type SSEFetchClientEvents = {
  message: SSEMessage;
  error: Error;
  close: void;
  open: void;
};

// ===== SSE FETCH CLIENT =====

/**
 * A generic, reusable client for handling Server-Sent Events (SSE) streams.
 *
 * This client handles the SSE protocol (parsing events, data, etc.) but does NOT
 * interpret the semantic meaning of events. Consumers decide how to handle
 * different event types and data formats.
 */
export class SSEFetchClient {
  private abortController: AbortController | undefined;

  private readonly events = new EventEmitter<SSEFetchClientEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;
  public readonly once = this.events.once.bind(
    this.events,
  ) as typeof this.events.once;

  /**
   * Start streaming from the given URL
   *
   * @param url - The SSE endpoint URL
   * @param options - Optional configuration
   */
  public async start(
    url: string,
    options: {
      method?: "GET" | "POST";
      body?: Record<string, unknown>;
      headers?: Record<string, string>;
      getJwt?: () => string | undefined;
    } = {},
  ): Promise<void> {
    // Clean up any existing connection
    this.stop();

    const {
      method = "GET",
      body,
      headers: customHeaders = {},
      getJwt,
    } = options;

    // Prepare headers with authentication
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...customHeaders,
    };

    const jwt = getJwt?.();
    if (jwt) {
      headers["Authorization"] = `Bearer ${jwt}`;
    }

    try {
      // Create abort controller for cancellation
      this.abortController = new AbortController();

      // Make the fetch request
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

      this.events.emit("open");

      // Process the SSE stream
      await this.processSSEStream(response.body);
    } catch (error) {
      if (error.name !== "AbortError") {
        const errorArg =
          error instanceof Error ? error : new Error(String(error));
        this.events.emit("error", errorArg);
      }
    } finally {
      this.stop();
      this.events.emit("close");
    }
  }

  /**
   * Stop the current streaming session
   */
  public stop(): void {
    if (this.abortController) {
      this.abortController.abort("SSE stream stopped by client");
      this.abortController = undefined;
    }
  }

  /**
   * Stop streaming and clear all event listeners
   * Call this when the client is no longer needed to prevent memory leaks
   */
  public cleanup(): void {
    this.stop();

    // Clear all event listeners
    this.events.clearListeners();
  }

  /**
   * Check if currently streaming
   */
  public isStreaming(): boolean {
    return this.abortController !== undefined;
  }

  /**
   * Process the SSE stream from the response body
   */
  private async processSSEStream(
    body: ReadableStream<Uint8Array>,
  ): Promise<void> {
    const reader = body.getReader();
    const decoder = new TextDecoder();
    let buffer = "";
    let currentEvent: Partial<SSEMessage> = {};

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        // Decode chunk and add to buffer
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");

        // Keep the last incomplete line in the buffer
        buffer = lines.pop() || "";

        // Process each complete line
        for (const line of lines) {
          if (isEventComplete(line)) {
            // Empty line signals end of event - emit if valid
            if (isValidEvent(currentEvent)) {
              this.events.emit("message", currentEvent);
            }
            currentEvent = {};
          } else {
            // Parse line and accumulate into current event
            parseSSELine(line, currentEvent);
          }
        }
      }

      // Emit any remaining event in the buffer
      if (isValidEvent(currentEvent)) {
        this.events.emit("message", currentEvent);
      }
    } finally {
      reader.releaseLock();
    }
  }
}
