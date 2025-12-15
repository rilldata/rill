import { Throttler } from "@rilldata/web-common/lib/throttler";
import { streamingFetchWrapper } from "@rilldata/web-common/runtime-client/fetch-streaming-wrapper";
import type {
  V1WatchFilesResponse,
  V1WatchLogsResponse,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client/index";
import { get, writable } from "svelte/store";
import { asyncWait } from "../lib/waitUtils";
import { runtime } from "./runtime-store";

export const MAX_RETRIES = 5;
const BACKOFF_DELAY = 1000;
const RETRY_COUNT_DELAY = 1000;
const RECONNECT_CALLBACK_DELAY = 150;

type WatchResponse =
  | V1WatchFilesResponse
  | V1WatchResourcesResponse
  | V1WatchLogsResponse;

type StreamingFetchResponse<Res extends WatchResponse> = {
  result?: Res;
  error?: { code: number; message: string };
};

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
  private controller: AbortController | undefined;
  private stream: AsyncGenerator<StreamingFetchResponse<Res>> | undefined;
  private eventSource: EventSource | undefined;
  private useSSE: boolean = false;
  private outOfFocusThrottler = new Throttler(120000, 20000);
  public retryAttempts = writable(0);
  private reconnectTimeout: ReturnType<typeof setTimeout> | undefined;
  private retryTimeout: ReturnType<typeof setTimeout> | undefined;
  private listeners: Listeners<Res> = new Map([
    ["response", []],
    ["reconnect", []],
  ]);
  public closed = writable(false);

  constructor(useSSE: boolean = false) {
    this.useSSE = useSSE;
  }

  public on<K extends keyof EventMap<Res>>(
    event: K,
    listener: Callback<Res, K>,
  ) {
    this.listeners.get(event)?.push(listener);
  }

  public heartbeat = async () => {
    if (get(this.closed)) {
      await this.reconnect();
    }
    this.throttle();
  };

  public async watch(url: string, useSSE: boolean = false) {
    this.cancel();
    this.url = url;

    if (useSSE !== undefined) {
      this.useSSE = useSSE;
    }

    if (this.useSSE) {
      this.listenSSE();
    } else {
      await this.listen();
    }

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

    // The stream was not cancelled, so don't reconnect
    if (
      (this.controller && !this.controller.signal.aborted) ||
      (this.eventSource && this.eventSource.readyState !== EventSource.CLOSED)
    )
      return;

    const currentAttempts = get(this.retryAttempts);

    if (currentAttempts >= MAX_RETRIES) throw new Error("Max retries exceeded");

    if (currentAttempts > 0) {
      const delay = BACKOFF_DELAY * 2 ** currentAttempts;
      await asyncWait(delay);
    }

    this.retryAttempts.update((n) => n + 1);

    if (this.useSSE) {
      this.listenSSE(true);
    } else {
      await this.listen(true);
    }
  }

  private cancel() {
    // Clean up fetch stream
    this.controller?.abort("Watch request cancelled");
    this.stream = this.controller = undefined;

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
      if (reconnect) {
        this.reconnectTimeout = setTimeout(() => {
          this.listeners.get("reconnect")?.forEach((cb) => void cb());
        }, RECONNECT_CALLBACK_DELAY);
      }

      this.closed.set(false);

      this.eventSource = new EventSource(sseUrl.toString());

      this.eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          this.listeners.get("response")?.forEach((cb) => void cb(data));
        } catch (error) {
          console.error("Error parsing SSE message:", error);
        }
      };

      this.eventSource.onerror = async () => {
        clearTimeout(this.retryTimeout);
        this.cancel();

        if (get(this.closed)) return;

        await this.reconnect();
      };

      this.retryTimeout = setTimeout(() => {
        this.retryAttempts.set(0);
      }, RETRY_COUNT_DELAY);
    } catch {
      clearTimeout(this.retryTimeout);
      this.cancel();
    }
  }

  private async listen(reconnect = false) {
    clearTimeout(this.reconnectTimeout);

    if (!this.url) throw new Error("URL not set");

    this.controller = new AbortController();
    this.stream = this.getFetchStream(this.url, this.controller);

    try {
      if (reconnect) {
        this.reconnectTimeout = setTimeout(() => {
          this.listeners.get("reconnect")?.forEach((cb) => void cb());
        }, RECONNECT_CALLBACK_DELAY);
      }

      this.closed.set(false);

      this.retryTimeout = setTimeout(() => {
        this.retryAttempts.set(0);
      }, RETRY_COUNT_DELAY);

      for await (const res of this.stream) {
        if (this.controller?.signal.aborted) break;
        if (res.error) throw new Error(res.error.message);

        if (res.result)
          this.listeners.get("response")?.forEach((cb) => void cb(res.result));
      }
    } catch {
      clearTimeout(this.retryTimeout);

      this.cancel();
      if (get(this.closed)) return;

      await this.reconnect();
    }
  }

  private getFetchStream(url: string, controller: AbortController) {
    const headers = { "Content-Type": "application/json" };
    const jwt = get(runtime).jwt;
    if (jwt) {
      headers["Authorization"] = `Bearer ${jwt.token}`;
    }

    return streamingFetchWrapper<StreamingFetchResponse<Res>>(
      url,
      "GET",
      undefined,
      headers,
      controller.signal,
    );
  }
}
