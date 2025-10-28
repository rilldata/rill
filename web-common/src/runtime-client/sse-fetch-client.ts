import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

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
  private listeners: {
    message: ((message: SSEMessage) => void)[];
    error: ((error: Error) => void)[];
    close: (() => void)[];
  } = {
    message: [],
    error: [],
    close: [],
  };

  /**
   * Add event listener for SSE events
   */
  public on(event: "message", listener: (message: SSEMessage) => void): void;
  public on(event: "error", listener: (error: Error) => void): void;
  public on(event: "close", listener: () => void): void;
  public on(event: string, listener: any): void {
    if (this.listeners[event as keyof typeof this.listeners]) {
      this.listeners[event as keyof typeof this.listeners].push(listener);
    }
  }

  /**
   * Remove event listener
   */
  public off(event: "message", listener: (message: SSEMessage) => void): void;
  public off(event: "error", listener: (error: Error) => void): void;
  public off(event: "close", listener: () => void): void;
  public off(event: string, listener: any): void {
    const eventListeners = this.listeners[event as keyof typeof this.listeners];
    if (eventListeners) {
      const index = eventListeners.indexOf(listener);
      if (index > -1) {
        eventListeners.splice(index, 1);
      }
    }
  }

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
    } = {},
  ): Promise<void> {
    // Clean up any existing connection
    this.stop();

    const { method = "GET", body, headers: customHeaders = {} } = options;

    // Prepare headers with authentication
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      ...customHeaders,
    };

    const jwt = get(runtime).jwt;
    if (jwt) {
      headers["Authorization"] = `Bearer ${jwt.token}`;
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

      // Process the SSE stream
      await this.processSSEStream(response.body);
    } catch (error) {
      if (error.name !== "AbortError") {
        this.listeners.error.forEach((listener) =>
          listener(error instanceof Error ? error : new Error(String(error))),
        );
      }
    } finally {
      this.listeners.close.forEach((listener) => listener());
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
    this.listeners.message = [];
    this.listeners.error = [];
    this.listeners.close = [];
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
              this.emitMessage(currentEvent);
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
        this.emitMessage(currentEvent);
      }
    } finally {
      reader.releaseLock();
    }
  }

  /**
   * Emit a message to all registered listeners
   */
  private emitMessage(message: SSEMessage): void {
    this.listeners.message.forEach((listener) => listener(message));
  }
}
