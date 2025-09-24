import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

/**
 * A clean, reusable client for handling Server-Sent Events (SSE) streams.
 *
 * This client properly handles the SSE protocol with "data: " prefixes and provides
 * a simple event-based interface for consuming streaming responses.
 */
export class SSEFetchClient<T> {
  private abortController: AbortController | undefined;
  private listeners: {
    data: ((data: T) => void)[];
    error: ((error: Error) => void)[];
    close: (() => void)[];
  } = {
    data: [],
    error: [],
    close: [],
  };

  constructor(private readonly options?: { includeAuth?: boolean }) {
    // Default to including auth
    this.options = { includeAuth: true, ...options };
  }

  /**
   * Add event listener for SSE events
   */
  public on(event: "data", listener: (data: T) => void): void;
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
  public off(event: "data", listener: (data: T) => void): void;
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

    // Only include auth if explicitly requested
    if (this.options?.includeAuth) {
      const jwt = get(runtime).jwt;
      if (jwt) {
        headers["Authorization"] = `Bearer ${jwt.token}`;
      }
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
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
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
      this.abortController.abort();
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
    this.listeners.data = [];
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

    try {
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;

        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split("\n");

        // Keep the last incomplete line in the buffer
        buffer = lines.pop() || "";

        for (const line of lines) {
          await this.processSSELine(line.trim());
        }
      }

      // Process any remaining data in the buffer
      if (buffer.trim()) {
        await this.processSSELine(buffer.trim());
      }
    } finally {
      reader.releaseLock();
    }
  }

  /**
   * Process a single SSE line
   */
  private async processSSELine(line: string): Promise<void> {
    // Skip empty lines and comments
    if (!line || line.startsWith(":")) {
      return;
    }

    // Handle SSE data lines
    if (line.startsWith("data: ")) {
      try {
        const jsonData = line.slice(6); // Remove "data: " prefix
        const data: T = JSON.parse(jsonData);
        this.listeners.data.forEach((listener) => listener(data));
      } catch (error) {
        this.listeners.error.forEach((listener) =>
          listener(new Error(`Failed to parse SSE data: ${error.message}`)),
        );
      }
    }

    // Note: We could extend this to handle other SSE fields like "event:", "id:", "retry:" if needed
  }
}
