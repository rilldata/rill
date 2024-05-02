import { Throttler } from "@rilldata/web-common/lib/throttler";
import { ExponentialBackoffTracker } from "@rilldata/web-common/runtime-client/exponential-backoff-tracker";
import { streamingFetchWrapper } from "@rilldata/web-common/runtime-client/fetch-streaming-wrapper";
import type {
  V1WatchFilesResponse,
  V1WatchLogsResponse,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client/index";
import { get } from "svelte/store";
import { runtime } from "./runtime-store";

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
  private tracker = ExponentialBackoffTracker.createBasicTracker();
  private outOfFocusThrottler = new Throttler(10000);
  private listeners: Listeners<Res> = new Map([
    ["response", []],
    ["reconnect", []],
  ]);

  on<K extends keyof EventMap<Res>>(event: K, listener: Callback<Res, K>) {
    this.listeners.get(event)?.push(listener);
  }

  watch(url: string) {
    this.cancel();
    this.url = url;
    this.init();
    this.listen().catch(console.error);
  }

  cancel() {
    this.controller?.abort();
    this.stream = this.controller = undefined;
  }

  init() {
    if (!this.url) throw new Error("URL not set");
    this.controller = new AbortController();
    this.stream = this.getFetchStream(this.url, this.controller);
  }

  throttle() {
    this.outOfFocusThrottler.throttle(() => {
      this.cancel();
    });
  }

  reconnect() {
    if (this.outOfFocusThrottler.isThrottling()) {
      this.outOfFocusThrottler.cancel();
    }

    // The stream was not cancelled, so don't reconnect
    if (this.controller && !this.controller.signal.aborted) return;

    this.init();
    this.listen().catch(console.error);

    // Reconnecting, notify listeners
    this.listeners.get("reconnect")?.forEach((cb) => void cb());
  }

  private async listen() {
    if (!this.stream) return;
    try {
      for await (const res of this.stream) {
        if (this.controller?.signal.aborted) break;
        if (res.error) throw new Error(res.error.message);

        if (res.result)
          this.listeners.get("response")?.forEach((cb) => void cb(res.result));
      }
    } catch (err) {
      // Stream failed, attempt to reconnect with exponential backoff
      this.controller = undefined;
      this.tracker.try(() => this.reconnect());
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
