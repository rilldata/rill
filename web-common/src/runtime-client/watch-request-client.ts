import { Throttler } from "@rilldata/web-common/lib/throttler";
import type {
  V1WatchFilesResponse,
  V1WatchLogsResponse,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client/index";
import { get, writable } from "svelte/store";
import { asyncWait } from "../lib/waitUtils";
import { SSEFetchClient } from "./sse-fetch-client";

// Retry configuration
const MAX_RETRIES = 5;
const BACKOFF_DELAY = 1000; // Base delay in ms
const RETRY_COUNT_DELAY = 500; // Delay before resetting retry count
const RECONNECT_CALLBACK_DELAY = 150; // Delay before firing reconnect callbacks

// Throttling configuration
const OUT_OF_FOCUS_TIMEOUT = 120000; // 2 minutes
const OUT_OF_FOCUS_SHORT_TIMEOUT = 20000; // 20 seconds

type WatchResponse =
  | V1WatchFilesResponse
  | V1WatchResourcesResponse
  | V1WatchLogsResponse;

type EventMap<T> = {
  response: T;
  reconnect: void;
};

type Listeners<T> = Map<keyof EventMap<T>, Callback<T, keyof EventMap<T>>[]>;

type Callback<T, K extends keyof EventMap<T>> = (
  eventData: EventMap<T>[K],
) => void | Promise<void>;

interface WatchRequestClientOptions {
  includeAuth?: boolean;
}

/**
 * A wrapper around SSEFetchClient that adds watch-specific functionality:
 *
 * - Retry logic with exponential backoff (up to 5 attempts)
 * - Out-of-focus throttling (closes connections when page is not visible)
 * - Automatic reconnection management with callback notifications
 * - Event mapping (SSE "data" events → "response" events for semantic clarity)
 *
 * TODO: Consider moving retry/throttling functionality into SSEFetchClient itself
 * to avoid this additional abstraction layer. Currently this wrapper exists because
 * the watch clients have specific retry and lifecycle requirements that differ
 * from the simpler streaming needs of other SSE consumers like the chat system.
 */
export class WatchRequestClient<Res extends WatchResponse> {
  private url: string | undefined;
  private sseClient: SSEFetchClient<Res> | undefined;
  private outOfFocusThrottler = new Throttler(
    OUT_OF_FOCUS_TIMEOUT,
    OUT_OF_FOCUS_SHORT_TIMEOUT,
  );
  public retryAttempts = writable(0);
  private reconnectTimeout: ReturnType<typeof setTimeout> | undefined;
  private retryTimeout: ReturnType<typeof setTimeout> | undefined;
  private listeners: Listeners<Res> = new Map([
    ["response", []],
    ["reconnect", []],
  ]);
  public closed = writable(false);
  private isReconnecting = false;

  constructor(private readonly options?: WatchRequestClientOptions) {
    // Default to no auth
    this.options = { includeAuth: false, ...options };
  }

  // ===== PUBLIC API =====

  /**
   * Register an event listener
   */
  public on<K extends keyof EventMap<Res>>(
    event: K,
    listener: Callback<Res, K>,
  ) {
    this.listeners.get(event)?.push(listener);
  }

  /**
   * Start watching a URL for SSE events
   */
  public watch(url: string) {
    this.cancel();
    this.url = url;

    void this.startSSE();

    // Enable auto-close behavior to manage HTTP connection quota (browsers limit ~6 concurrent connections per host)
    this.throttle();
  }

  /**
   * Keep the connection alive (called on user interaction)
   */
  public heartbeat = () => {
    if (get(this.closed)) {
      void this.reconnect().catch((e) => {
        console.error("Reconnection failed:", e);
      });
    }
    // Reset auto-close timer to keep this tab's connection active
    this.throttle();
  };

  /**
   * Enable auto-close behavior (closes connection after period of inactivity)
   *
   * Note: The name 'throttle' is historical - this actually starts a timer to close the connection
   */
  public throttle(prioritize: boolean = false) {
    this.outOfFocusThrottler.cancel();
    this.outOfFocusThrottler.throttle(this.close, prioritize);
  }

  /**
   * Close the connection and clean up all resources
   */
  public close = () => {
    this.cancel();
    this.closed.set(true);

    // Clean up SSE client completely when closing
    if (this.sseClient) {
      this.sseClient.cleanup();
      this.sseClient = undefined;
    }
  };

  // ===== PRIVATE CORE FUNCTIONALITY =====

  /**
   * Start the SSE connection
   */
  private async startSSE() {
    if (!this.url) throw new Error("URL not set");

    try {
      this.closed.set(false);

      // Always create a fresh SSE client to avoid connection reuse issues
      if (this.sseClient) {
        this.sseClient.cleanup();
      }

      this.sseClient = new SSEFetchClient<Res>(this.options);

      // Set up event handlers for the new client
      this.setupSSEEventHandlers();

      // Start streaming
      await this.sseClient.start(this.url);

      // Only set up retry timeout reset if the connection was successful
      this.retryTimeout = setTimeout(() => {
        this.retryAttempts.set(0);
      }, RETRY_COUNT_DELAY);
    } catch (error) {
      console.error("Failed to start SSE:", error);
      this.handleError();
    }
  }

  /**
   * Handle reconnection with exponential backoff
   */
  private async reconnect() {
    // Prevent concurrent reconnection attempts
    if (this.isReconnecting) return;
    this.isReconnecting = true;

    try {
      clearTimeout(this.reconnectTimeout);

      if (this.outOfFocusThrottler.isThrottling()) {
        this.outOfFocusThrottler.cancel();
      }

      // Don't reconnect if client is already streaming
      if (this.sseClient?.isStreaming()) {
        return;
      }

      const currentAttempts = get(this.retryAttempts);

      if (currentAttempts >= MAX_RETRIES) {
        throw new Error("Max retries exceeded");
      }

      if (currentAttempts > 0) {
        const delay = BACKOFF_DELAY * 2 ** currentAttempts;
        await asyncWait(delay);
      }

      this.retryAttempts.update((n) => n + 1);

      // Fire reconnect callbacks after a short delay
      this.reconnectTimeout = setTimeout(() => {
        this.fireCallbacks("reconnect", undefined);
      }, RECONNECT_CALLBACK_DELAY);

      void this.startSSE();
    } finally {
      this.isReconnecting = false;
    }
  }

  /**
   * Cancel the current connection and clean up timeouts
   */
  private cancel() {
    // Clean up SSE connection
    if (this.sseClient) {
      this.sseClient.stop();
    }

    // Clear timeouts
    clearTimeout(this.reconnectTimeout);
    clearTimeout(this.retryTimeout);
  }

  /**
   * Handle errors by canceling and attempting to reconnect
   */
  private handleError() {
    this.cancel();
    if (get(this.closed)) return;

    void this.reconnect().catch((e) => {
      console.error("Reconnection failed:", e);
    });
  }

  // ===== PRIVATE HELPER METHODS =====

  /**
   * Set up event handlers for the SSE client
   */
  private setupSSEEventHandlers() {
    if (!this.sseClient) return;

    this.sseClient.on("data", (data) => {
      // Map SSE "data" events to "response" events for semantic clarity
      this.fireCallbacks("response", data);
    });

    this.sseClient.on("error", (error) => {
      console.error("SSE error:", error);
      this.handleError();
    });

    this.sseClient.on("close", () => {
      // No action
    });
  }

  /**
   * Fire callbacks for a specific event
   */
  private fireCallbacks<K extends keyof EventMap<Res>>(
    event: K,
    data: EventMap<Res>[K],
  ) {
    this.listeners.get(event)?.forEach((cb) => void cb(data));
  }
}
