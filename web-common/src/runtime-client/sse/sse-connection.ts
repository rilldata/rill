import { createEventBinding } from "@rilldata/web-common/lib/event-emitter.ts";
import { get, writable } from "svelte/store";
import { SSEConnectionLifecycle } from "./sse-connection-lifecycle";
import { SSEFetchClient, type SSEMessage } from "./sse-fetch-client";

const BACKOFF_DELAY = 1000;
// A connection must stay open this long before it's considered "stable"
// enough that a subsequent failure should start with a fresh retry budget.
const MIN_STABLE_DURATION = 5000;

export type SSEConnectionOptions = {
  /**
   * @deprecated Pass an SSEConnectionLifecycle alongside the connection instead.
   * Retained so existing consumers compile unchanged; removed in the
   * consumer-migration PR.
   */
  autoCloseTimeouts?: {
    short: number;
    normal: number;
  };
  maxRetryAttempts?: number;
  retryOnError?: boolean;
  retryOnClose?: boolean;
  /**
   * Hook fired after the backoff delay and before each reconnect attempt's
   * transport call. Lets long-lived streams refresh auth (e.g. the cloud
   * editor's JWT) before opening a new connection.
   *
   * If the hook rejects, the transport call is skipped for that attempt, an
   * `error` event is emitted, and the rejection counts as a failed attempt
   * — so repeated hook failures walk the status to CLOSED the same way
   * repeated transport failures do.
   */
  onBeforeReconnect?: () => Promise<void>;
};

export type SSEStartOptions = {
  method?: "GET" | "POST";
  body?: Record<string, unknown>;
  headers?: Record<string, string>;
  getJwt?: () => string | undefined;
};

export enum ConnectionStatus {
  CONNECTING = "connecting",
  OPEN = "open",
  PAUSED = "paused",
  CLOSED = "closed",
}

type SSEConnectionEvents = {
  message: SSEMessage;
  reconnect: void;
  error: Error;
  close: void;
  open: void;
};

/**
 * Transport-control layer over SSEFetchClient. Owns status
 * (CONNECTING / OPEN / PAUSED / CLOSED), exponential-backoff retry up to
 * maxRetryAttempts, retry-count reset after a stable connection, and an
 * optional onBeforeReconnect hook for auth refresh between retries. Emits
 * `reconnect` only on post-initial successful opens.
 *
 * Does not decide lifecycle policy (for example pausing on tab hide or
 * resuming on activity). Attach `SSEConnectionLifecycle` for that.
 */
export class SSEConnection {
  public status = writable<ConnectionStatus>(ConnectionStatus.CLOSED);

  public url: string;
  public options: SSEStartOptions;

  private readonly events = createEventBinding<SSEConnectionEvents>();
  public readonly on = this.events.on;
  public readonly once = this.events.once;

  private client = new SSEFetchClient();

  /**
   * @deprecated Legacy auto-close compatibility shim. PR 2 removes these
   * forwarding paths in favor of attaching SSEConnectionLifecycle directly.
   */
  private autoCloseLifecycle: SSEConnectionLifecycle | undefined;
  private autoCloseDisabled = false;

  private retryAttemptCount = writable(0);
  // Single-flight guard + cancellation signal for reconnect().
  // When set, one reconnect loop is active; calling abortReconnect() aborts
  // backoff/hook work and guarantees that loop exits before openConnection().
  private reconnectController: AbortController | null = null;
  private connectionCount = 0;
  private openedAt: number | null = null;

  constructor(public params?: SSEConnectionOptions) {
    if (params?.autoCloseTimeouts) {
      this.autoCloseLifecycle = new SSEConnectionLifecycle(
        this,
        params.autoCloseTimeouts,
      );
    }

    this.client.on("error", this.handleError);
    this.client.on("message", this.handleMessage);
    this.client.on("close", this.handleCloseEvent);
    this.client.on("open", this.handleSuccessfulConnection);
  }

  /**
   * Start streaming from the given URL. Begins a new logical session, so
   * retry state is cleared.
   */
  public start(url: string, options: SSEStartOptions = {}): void {
    this.url = url;
    this.options = options;
    this.retryAttemptCount.set(0);
    this.connectionCount = 0;
    this.openedAt = null;
    this.abortReconnect();
    this.openConnection();
  }

  /**
   * Resume from PAUSED if necessary, and re-arm auto-close for legacy
   * compatibility paths that still call scheduleAutoClose/heartbeat.
   */
  public resumeIfPaused = async () => {
    if (get(this.status) === ConnectionStatus.PAUSED) {
      await this.reconnect();
    }

    if (this.autoCloseLifecycle) {
      this.scheduleAutoClose();
    }
  };

  /**
   * @deprecated Use resumeIfPaused() instead.
   */
  public heartbeat = async () => {
    await this.resumeIfPaused();
  };

  /**
   * Pause the transport. Preserves reconnect ability via resumeIfPaused(),
   * unlike close() which terminates the session.
   */
  public pause(): void {
    const status = get(this.status);
    if (
      status === ConnectionStatus.CLOSED ||
      status === ConnectionStatus.PAUSED
    )
      return;

    // Intentional pause isn't a failure, so reset the retry budget.
    this.retryAttemptCount.set(0);
    this.openedAt = null;
    this.abortReconnect();
    this.status.set(ConnectionStatus.PAUSED);
    this.client.stop();
  }

  /**
   * Terminate the session. Pass cleanup=true to also clear listeners.
   */
  public close = (cleanup = false) => {
    if (get(this.status) === ConnectionStatus.CLOSED) {
      if (cleanup) this.events.clearListeners();
      return;
    }

    this.openedAt = null;
    this.abortReconnect();
    this.status.set(ConnectionStatus.CLOSED);
    this.client.stop();
    this.events.emit("close");

    if (cleanup) {
      this.events.clearListeners();
    }
  };

  /**
   * Fully tear down the session and listeners.
   */
  public cleanup(): void {
    this.close(true);
  }

  /**
   * @deprecated Move to SSEConnectionLifecycle; removed in the
   * consumer-migration PR.
   */
  public scheduleAutoClose(prioritize: boolean = false) {
    if (this.autoCloseDisabled) return;
    this.autoCloseLifecycle?.schedulePause(prioritize);
  }

  /**
   * @deprecated Move to SSEConnectionLifecycle; removed in the
   * consumer-migration PR.
   */
  public disableAutoClose() {
    this.autoCloseDisabled = true;
    this.autoCloseLifecycle?.cancelScheduledPause();
  }

  /**
   * @deprecated Move to SSEConnectionLifecycle; removed in the
   * consumer-migration PR.
   */
  public enableAutoClose() {
    this.autoCloseDisabled = false;
  }

  private openConnection(): void {
    this.status.set(ConnectionStatus.CONNECTING);
    void this.client.start(this.url, this.options);
    if (this.autoCloseLifecycle) {
      this.scheduleAutoClose();
    }
  }

  private async reconnect() {
    if (this.reconnectController) return;
    const controller = new AbortController();
    this.reconnectController = controller;
    const { signal } = controller;

    try {
      while (!signal.aborted) {
        this.autoCloseLifecycle?.cancelScheduledPause();

        const status = get(this.status);
        if (
          status === ConnectionStatus.OPEN ||
          status === ConnectionStatus.CLOSED
        )
          return;

        const currentAttempts = get(this.retryAttemptCount);
        if (currentAttempts >= (this.params?.maxRetryAttempts ?? 0)) {
          this.close();
          return;
        }

        if (currentAttempts > 0) {
          await waitOrAbort(BACKOFF_DELAY * 2 ** currentAttempts, signal);
          if (signal.aborted) return;
        }

        this.retryAttemptCount.update((n) => n + 1);

        if (this.params?.onBeforeReconnect) {
          try {
            await this.params.onBeforeReconnect();
          } catch (err) {
            const errorArg =
              err instanceof Error ? err : new Error(String(err));
            this.events.emit("error", errorArg);
            // Hook failure counts as a failed attempt; stay in-loop under
            // the same guard so the next iteration applies backoff.
            continue;
          }
          if (signal.aborted) return;
        }

        this.openConnection();
        return;
      }
    } finally {
      if (this.reconnectController === controller) {
        this.reconnectController = null;
      }
    }
  }

  private abortReconnect(): void {
    this.reconnectController?.abort();
    this.reconnectController = null;
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
    } else {
      // No retry configured: terminate the session and fire `close` so
      // awaiters (e.g. one-shot chat streams) can settle.
      this.close();
    }

    const errorArg = error instanceof Error ? error : new Error(String(error));
    this.events.emit("error", errorArg);
  };

  // Fires when the underlying fetch ends. Client-initiated closes (pause()
  // or close()) set status away from OPEN before stopping the client, so
  // the guard below filters them out; this handler runs only for
  // server-initiated closes.
  private handleCloseEvent = () => {
    if (get(this.status) !== ConnectionStatus.OPEN) return;

    // Reset retries when a stable connection closes so the next failure
    // starts with a fresh budget; unstable closes (server opens then
    // immediately closes) still accumulate attempts.
    const wasStable =
      this.openedAt !== null &&
      Date.now() - this.openedAt >= MIN_STABLE_DURATION;
    if (wasStable) {
      this.retryAttemptCount.set(0);
    }
    this.openedAt = null;

    if (this.params?.retryOnClose) {
      this.status.set(ConnectionStatus.CONNECTING);
      void this.reconnect();
    } else {
      this.close();
    }
  };

  private handleMessage = (message: SSEMessage) => {
    this.events.emit("message", message);
  };

  private handleSuccessfulConnection = () => {
    this.connectionCount += 1;
    this.openedAt = Date.now();
    this.status.set(ConnectionStatus.OPEN);
    this.events.emit("open");

    // Mirror handleCloseEvent's wasStable reset for the case where the
    // connection later errors out instead of closing cleanly. A stale timer
    // from an earlier open no-ops here because `openedAt` reflects the
    // current connection — a rapid reopen overwrites it, and pause/close
    // clear it to null.
    setTimeout(() => {
      if (
        this.openedAt !== null &&
        Date.now() - this.openedAt >= MIN_STABLE_DURATION
      ) {
        this.retryAttemptCount.set(0);
      }
    }, MIN_STABLE_DURATION);

    if (this.connectionCount > 1) {
      this.events.emit("reconnect");
    }
  };
}

function waitOrAbort(ms: number, signal: AbortSignal): Promise<void> {
  if (signal.aborted) return Promise.resolve();
  return new Promise((resolve) => {
    const timer = setTimeout(() => {
      signal.removeEventListener("abort", onAbort);
      resolve();
    }, ms);
    const onAbort = () => {
      clearTimeout(timer);
      resolve();
    };
    signal.addEventListener("abort", onAbort, { once: true });
  });
}
