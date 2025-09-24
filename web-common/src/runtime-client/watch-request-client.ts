import { Throttler } from "@rilldata/web-common/lib/throttler";
import type {
  V1WatchFilesResponse,
  V1WatchLogsResponse,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client/index";
import { get, writable } from "svelte/store";
import { asyncWait } from "../lib/waitUtils";

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

export class WatchRequestClient<Res extends WatchResponse> {
  private url: string | undefined;
  private eventSource: EventSource | undefined;
  private outOfFocusThrottler = new Throttler(120000, 20000);
  public retryAttempts = writable(0);
  private reconnectTimeout: ReturnType<typeof setTimeout> | undefined;
  private retryTimeout: ReturnType<typeof setTimeout> | undefined;
  private listeners: Listeners<Res> = new Map([
    ["response", []],
    ["reconnect", []],
  ]);
  public closed = writable(false);

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

    this.listenSSE();

    // Start throttling after the first connection
    this.throttle();
  }

  public close = () => {
    this.cancel();
    this.closed.set(true);
  };

  public throttle(prioritize: boolean = false) {
    this.outOfFocusThrottler.cancel();
    this.outOfFocusThrottler.throttle(this.close, prioritize);
  }

  private async reconnect() {
    clearTimeout(this.reconnectTimeout);

    if (this.outOfFocusThrottler.isThrottling()) {
      this.outOfFocusThrottler.cancel();
    }

    // The SSE connection was not cancelled, so don't reconnect
    if (this.eventSource && this.eventSource.readyState !== EventSource.CLOSED)
      return;

    const currentAttempts = get(this.retryAttempts);

    if (currentAttempts >= MAX_RETRIES) throw new Error("Max retries exceeded");

    if (currentAttempts > 0) {
      const delay = BACKOFF_DELAY * 2 ** currentAttempts;
      await asyncWait(delay);
    }

    this.retryAttempts.update((n) => n + 1);

    this.listenSSE(true);
  }

  private cancel() {
    // Clean up SSE connection
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = undefined;
    }
  }

  private listenSSE(reconnect = false) {
    clearTimeout(this.reconnectTimeout);

    if (!this.url) throw new Error("URL not set");

    const sseUrl = new URL(this.url, window.location.origin);

    try {
      this.retryTimeout = setTimeout(() => {
        this.retryAttempts.set(0);
      }, RETRY_COUNT_DELAY);

      if (reconnect) {
        this.reconnectTimeout = setTimeout(() => {
          this.listeners.get("reconnect")?.forEach((cb) => void cb());
        }, RECONNECT_CALLBACK_DELAY);
      }

      this.closed.set(false);

      // Create new EventSource
      this.eventSource = new EventSource(sseUrl.toString());

      // Handle incoming messages
      this.eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.listeners.get("response")?.forEach((cb) => void cb(data));
        } catch (error) {
          console.error("Error parsing SSE message:", error);
        }
      };

      // Handle connection errors
      this.eventSource.onerror = () => {
        this.cancel();
        if (get(this.closed)) return;

        this.reconnect().catch((e) => {
          console.error("Reconnection failed:", e);
          // Or rethrow the original error if needed
          throw e;
        });
      };
    } catch {
      clearTimeout(this.retryTimeout);

      this.cancel();
      if (get(this.closed)) return;

      this.reconnect().catch((e) => {
        console.error("Reconnection failed:", e);
        // Or rethrow the original error if needed
        throw e;
      });
    }
  }
}
