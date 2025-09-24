import { Throttler } from "@rilldata/web-common/lib/throttler";
import type {
  V1WatchFilesResponse,
  V1WatchLogsResponse,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client/index";
import { get, writable } from "svelte/store";
import { asyncWait } from "../lib/waitUtils";
import { SSEFetchClient } from "./sse-fetch-client";

const MAX_RETRIES = 5;
const BACKOFF_DELAY = 1000;
const RETRY_COUNT_DELAY = 500;
const RECONNECT_CALLBACK_DELAY = 150;

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

/**
 * A wrapper around SSEFetchClient that adds watch-specific functionality:
 *
 * - Retry logic with exponential backoff (up to 5 attempts)
 * - Out-of-focus throttling (closes connections when page is not visible)
 * - Automatic reconnection management with callback notifications
 * - Event mapping (SSE "data" events → "response" events for backward compatibility)
 *
 * TODO: Consider moving retry/throttling functionality into SSEFetchClient itself
 * to avoid this additional abstraction layer. Currently this wrapper exists because
 * the watch clients have specific retry and lifecycle requirements that differ
 * from the simpler streaming needs of other SSE consumers like the chat system.
 */
export class WatchRequestClient<Res extends WatchResponse> {
  private url: string | undefined;
  private sseClient: SSEFetchClient<Res> | undefined;
  private outOfFocusThrottler = new Throttler(120000, 20000);
  public retryAttempts = writable(0);
  private reconnectTimeout: ReturnType<typeof setTimeout> | undefined;
  private retryTimeout: ReturnType<typeof setTimeout> | undefined;
  private listeners: Listeners<Res> = new Map([
    ["response", []],
    ["reconnect", []],
  ]);
  public closed = writable(false);
  private isReconnecting = false;

  constructor(private readonly options?: { includeAuth?: boolean }) {
    // Default to no auth for backward compatibility with local dev
    this.options = { includeAuth: false, ...options };
  }

  public on<K extends keyof EventMap<Res>>(
    event: K,
    listener: Callback<Res, K>,
  ) {
    this.listeners.get(event)?.push(listener);
  }

  public heartbeat = () => {
    if (get(this.closed)) {
      this.reconnect().catch((e) => {
        console.error("Reconnection failed:", e);
        throw e;
      });
    }
    this.throttle();
  };

  public watch(url: string) {
    this.cancel();
    this.url = url;

    void this.startSSE();

    // Start throttling after the first connection
    this.throttle();
  }

  public close = () => {
    this.cancel();
    this.closed.set(true);

    // Clean up SSE client completely when closing
    if (this.sseClient) {
      this.sseClient.cleanup();
      this.sseClient = undefined;
    }
  };

  public throttle(prioritize: boolean = false) {
    this.outOfFocusThrottler.cancel();
    this.outOfFocusThrottler.throttle(this.close, prioritize);
  }

  private async reconnect() {
    // Prevent concurrent reconnection attempts
    if (this.isReconnecting) return;
    this.isReconnecting = true;

    try {
      clearTimeout(this.reconnectTimeout);

      if (this.outOfFocusThrottler.isThrottling()) {
        this.outOfFocusThrottler.cancel();
      }

      // Don't reconnect if client is streaming or if we're closed
      if (this.sseClient?.isStreaming() || get(this.closed)) {
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

      void this.startSSE(true);
    } finally {
      this.isReconnecting = false;
    }
  }

  private cancel() {
    // Clean up SSE connection
    if (this.sseClient) {
      this.sseClient.stop();
    }

    // Clear timeouts
    clearTimeout(this.reconnectTimeout);
    clearTimeout(this.retryTimeout);
  }

  private async startSSE(reconnect = false) {
    clearTimeout(this.reconnectTimeout);

    if (!this.url) throw new Error("URL not set");

    try {
      if (reconnect) {
        this.reconnectTimeout = setTimeout(() => {
          this.listeners.get("reconnect")?.forEach((cb) => void cb());
        }, RECONNECT_CALLBACK_DELAY);
      }

      this.closed.set(false);

      // Always create a fresh SSE client to avoid connection reuse issues
      if (this.sseClient) {
        this.sseClient.cleanup();
      }

      this.sseClient = new SSEFetchClient<Res>(this.options);

      // Set up event handlers for the new client
      this.sseClient.on("data", (data) => {
        this.listeners.get("response")?.forEach((cb) => void cb(data));
      });

      this.sseClient.on("error", (error) => {
        console.error("SSE error:", error);
        this.handleError();
      });

      this.sseClient.on("close", () => {
        // Connection closed - attempt reconnect if not intentionally closed
        if (!get(this.closed)) {
          this.handleError();
        }
      });

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

  private handleError() {
    this.cancel();
    if (get(this.closed)) return;

    this.reconnect().catch((e) => {
      console.error("Reconnection failed:", e);
    });
  }
}
