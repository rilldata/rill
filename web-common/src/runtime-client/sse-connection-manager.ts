import { get, writable } from "svelte/store";
import { Throttler } from "../lib/throttler";
import { asyncWait } from "../lib/waitUtils";
import { SSEFetchClient, type SSEMessage } from "./sse-fetch-client";
import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";

const BACKOFF_DELAY = 1000; // Base delay in ms

type Params = {
  autoCloseTimeouts?: {
    short: number;
    normal: number;
  };
  maxRetryAttempts?: number;
  retryOnError?: boolean;
  retryOnClose?: boolean;
};

export enum ConnectionStatus {
  CONNECTING = "connecting", // attempting to connect and has not yet hit max retries
  OPEN = "open", // actively streaming
  PAUSED = "paused", // disconnected, but can be restarted with a heartbeat
  CLOSED = "closed",
}

type SSEConnectionManagerEvents = {
  message: SSEMessage;
  reconnect: void;
  error: Error;
  close: void;
  open: void;
};

// ===== SSE CONNECTION MANAGER =====

/**
 * A wrapper around SSEFetchClient to manage status and reconnections
 */
export class SSEConnectionManager {
  public status = writable<ConnectionStatus>(ConnectionStatus.CLOSED);

  public url: string;
  public options: {
    method?: "GET" | "POST";
    body?: Record<string, unknown>;
    headers?: Record<string, string>;
  };

  private readonly events = new EventEmitter<SSEConnectionManagerEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;
  public readonly once = this.events.once.bind(
    this.events,
  ) as typeof this.events.once;

  private client = new SSEFetchClient();

  private autoCloseThrottler: Throttler | undefined;
  private retryAttempts = writable(0);
  private isReconnecting = false;
  private connectionCount = 0;

  constructor(public params?: Params) {
    if (params?.autoCloseTimeouts) {
      this.autoCloseThrottler = new Throttler(
        params.autoCloseTimeouts.normal,
        params.autoCloseTimeouts.short,
      );
    }

    this.client.on("error", this.handleError);
    this.client.on("message", this.handleMessage);
    this.client.on("close", this.handleCloseEvent);
    this.client.on("open", this.handleSuccessfulConnection);
  }

  /**
   * Handle reconnection with exponential backoff
   */
  private async reconnect() {
    // Prevent concurrent reconnection attempts
    if (this.isReconnecting) {
      console.log("[SSE] reconnect skipped - already reconnecting");
      return;
    }
    this.isReconnecting = true;
    console.log("[SSE] reconnect started");

    try {
      if (this.autoCloseThrottler?.isThrottling()) {
        this.autoCloseThrottler.cancel();
      }

      // Don't reconnect if client is already streaming
      if (get(this.status) === ConnectionStatus.OPEN) {
        console.log("[SSE] reconnect skipped - already OPEN");
        return;
      }

      const currentAttempts = get(this.retryAttempts);
      console.log(
        "[SSE] reconnect attempt",
        currentAttempts + 1,
        "of",
        this.params?.maxRetryAttempts ?? 0,
      );

      if (currentAttempts >= (this.params?.maxRetryAttempts ?? 0)) {
        console.log("[SSE] MAX RETRIES EXCEEDED - setting CLOSED");
        this.status.set(ConnectionStatus.CLOSED);
        return;
      }

      if (currentAttempts > 0) {
        const delay = BACKOFF_DELAY * 2 ** currentAttempts;
        await asyncWait(delay);
      }

      this.retryAttempts.update((n) => n + 1);

      void this.start(this.url, this.options);
    } finally {
      this.isReconnecting = false;
      console.log("[SSE] reconnect finished, isReconnecting reset to false");
    }
  }

  /**
   * Stop the connection, mark closed and clean up resources
   */
  public heartbeat = async () => {
    const status = get(this.status);
    console.log("[SSE] heartbeat called, status:", status);
    // Only reconnect if PAUSED (intentionally disconnected to save resources)
    // Don't reconnect if CONNECTING (already in progress) or CLOSED (fatal error)
    if (status === ConnectionStatus.PAUSED) {
      console.log(
        "[SSE] heartbeat triggering reconnect because status is PAUSED",
      );
      await this.reconnect();
    }

    if (this.params?.autoCloseTimeouts) {
      this.scheduleAutoClose();
    }
  };

  /**
   * Stop the connection, mark closed and clean up resources
   */
  public close = (cleanup = false) => {
    this.pause();

    this.status.set(ConnectionStatus.CLOSED);

    if (cleanup) {
      this.cleanup();
    }
  };

  /**
   * Enable auto-close behavior to manage HTTP connection quota (browsers limit ~6 concurrent connections per host)
   */
  public scheduleAutoClose(prioritize: boolean = false) {
    this.autoCloseThrottler?.cancel();
    this.autoCloseThrottler?.throttle(() => this.pause(), prioritize);
  }

  private handleError = (error: Error) => {
    const status = get(this.status);

    if (
      status === ConnectionStatus.CLOSED ||
      status === ConnectionStatus.PAUSED
    )
      return;

    if (this.params?.retryOnError) {
      this.status.set(ConnectionStatus.CONNECTING);
      void this.reconnect();
    }

    const errorArg = error instanceof Error ? error : new Error(String(error));
    this.events.emit("error", errorArg);
  };

  // This can happen in one of three situations:
  // 1. The connection was paused intentionally (AbortError)
  // 2. There has been a network error causing the connection to close (FetchError)
  // 3. The application was terminated
  private handleCloseEvent = () => {
    const status = get(this.status);

    if (status !== ConnectionStatus.OPEN) return;

    if (this.params?.retryOnClose) {
      this.status.set(ConnectionStatus.CONNECTING);
      void this.reconnect();
    } else {
      this.close();

      this.events.emit("close");
    }
  };

  private handleMessage = (message: SSEMessage) => {
    this.events.emit("message", message);
  };

  private handleSuccessfulConnection = () => {
    this.connectionCount += 1;
    this.status.set(ConnectionStatus.OPEN);

    this.retryAttempts.set(0);

    if (this.connectionCount > 1) {
      this.events.emit("reconnect");
    }
  };

  /**
   * Start streaming from the given URL
   *
   * @param url - The SSE endpoint URL
   * @param options - Optional configuration
   */
  public start(
    url: string,
    options: {
      method?: "GET" | "POST";
      body?: Record<string, unknown>;
      headers?: Record<string, string>;
    } = {},
  ): void {
    console.log(
      "[SSE] start() called, current status:",
      get(this.status),
      "isStreaming:",
      this.client.isStreaming(),
    );
    this.url = url;
    this.options = options;

    this.status.set(ConnectionStatus.CONNECTING);

    void this.client.start(url);
    console.log("[SSE] client.start() initiated");

    if (this.params?.autoCloseTimeouts) {
      this.scheduleAutoClose();
    }
  }

  /**
   * Stop the current streaming session
   */
  public pause(): void {
    const status = get(this.status);

    if (
      status === ConnectionStatus.CLOSED ||
      status === ConnectionStatus.PAUSED
    )
      return;

    this.status.set(ConnectionStatus.PAUSED);

    // This will trigger an AbortError event and subsequently a "close" event
    // Which we ignore based on the current status
    this.client.stop();
  }

  /**
   * Stop streaming and clear all event listeners
   * Call this when the client is no longer needed to prevent memory leaks
   */
  public cleanup(): void {
    this.pause();

    // Clear all event listeners
    this.events.clearListeners();
  }
}
