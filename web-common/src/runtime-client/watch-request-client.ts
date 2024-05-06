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
  private tracker = ExponentialBackoffTracker.createBasicTracker();
  private outOfFocusThrottler = new Throttler(10000);
  private listeners: Listeners<Res> = new Map([
    ["response", []],
    ["reconnect", []],
  ]);

  public on<K extends keyof EventMap<Res>>(
    event: K,
    listener: Callback<Res, K>,
  ) {
    this.listeners.get(event)?.push(listener);
  }

  public watch(url: string) {
    this.cancel();
    this.url = url;
    this.restart().catch(console.log);
  }

  public cancel() {
    this.controller?.abort();
    this.controller = undefined;
  }

  public throttle() {
    this.outOfFocusThrottler.throttle(() => {
      this.cancel();
    });
  }

  public reconnect() {
    if (this.outOfFocusThrottler.isThrottling()) {
      this.outOfFocusThrottler.cancel();
    }

    // The stream was not cancelled, so don't reconnect
    if (this.controller && !this.controller.signal.aborted) return;

    this.restart().catch(console.log);

    // Reconnecting, notify listeners
    this.emitReconnect();
  }

  /**
   * (re)starts the stream connection for watch request.
   * If there is a disconnect then it reconnects with exponential backoff.
   */
  private async restart() {
    if (!this.url) return console.error("Unable to reconnect without a URL.");
    // abort previous connections before starting a new one
    this.controller?.abort();
    this.controller = new AbortController();

    let firstRun = true;
    // Maintain the controller here to make sure we check `aborted` for the correct one.
    // Checking for `this.controller` might lead to edge cases where it has a newer controller.
    let controller = this.controller;
    while (!controller.signal.aborted) {
      if (!firstRun) {
        // Reconnecting, notify listeners
        this.emitReconnect();
        // safeguard to cancel the request if not already cancelled
        controller.abort();
        controller = new AbortController();
      }
      firstRun = false;

      this.controller = controller;
      try {
        const stream = this.getFetchStream(this.url, this.controller);
        for await (const res of stream) {
          if (controller.signal.aborted) return;
          if (res.error) throw new Error(res.error.message);

          if (res.result) {
            this.listeners
              .get("response")
              ?.forEach((cb) => void cb(res.result));
          }
        }
      } catch (err) {
        if (!(await this.tracker.failed())) {
          // No point in continuing retry once we have failed enough times
          return;
        }
      }
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

  private emitReconnect() {
    this.listeners.get("reconnect")?.forEach((cb) => void cb());
  }
}
