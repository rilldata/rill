import { createEventBinding } from "@rilldata/web-common/lib/event-emitter.ts";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
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
 * Reconnect layer over SSEFetchClient. Owns:
 *   - Exponential-backoff retry up to maxRetryAttempts.
 *   - Connection-status store (CONNECTING / OPEN / PAUSED / CLOSED).
 *   - Stable-threshold retry-count reset (both open-then-stable and
 *     server-closes-stable paths).
 *   - Optional onBeforeReconnect hook for auth refresh.
 *
 * Does not own lifecycle policy (visibility, idle pausing). Attach an
 * SSEConnectionLifecycle if you want that.
 */
export class SSEConnection {
  public status = writable<ConnectionStatus>(ConnectionStatus.CLOSED);

  public url: string;
  public options: {
    method?: "GET" | "POST";
    body?: Record<string, unknown>;
    headers?: Record<string, string>;
    getJwt?: () => string | undefined;
  };

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

  private retryAttempts = writable(0);
  private isReconnecting = false;
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
   * retry state is cleared — previous failures shouldn't prevent a new
   * endpoint from connecting.
   */
  public start(
    url: string,
    options: {
      method?: "GET" | "POST";
      body?: Record<string, unknown>;
      headers?: Record<string, string>;
      getJwt?: () => string | undefined;
    } = {},
  ): void {
    this.url = url;
    this.options = options;
    this.retryAttempts.set(0);
    this.openConnection();
  }

  /**
   * Resume from PAUSED if necessary, and re-arm auto-close for legacy
   * compatibility paths that still call scheduleAutoClose/heartbeat.
   */
  public resumeIfPaused = async () => {
    const status = get(this.status);
    if (status === ConnectionStatus.PAUSED) {
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

    // Client-initiated close (auto-close after idle): reset retries because
    // an intentional pause is not a connection failure.
    // See also: handleCloseEvent() resets retries for server-initiated closes.
    this.retryAttempts.set(0);

    this.status.set(ConnectionStatus.PAUSED);

    // This fires a "close" event on the SSEFetchClient, but handleCloseEvent
    // ignores it because status is already PAUSED.
    this.client.stop();
  }

  /**
   * Transition to CLOSED. Pass cleanup=true to also clear listeners.
   */
  public close = (cleanup = false) => {
    this.pause();

    this.status.set(ConnectionStatus.CLOSED);

    if (cleanup) {
      this.cleanup();
    }
  };

  /**
   * Stop streaming and clear all connection listeners.
   */
  public cleanup(): void {
    this.pause();
    this.events.clearListeners();
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

  /**
   * Issue the underlying fetch and arm auto-close. Called by both start()
   * (new session) and reconnect() (retry of the current session); only
   * start() resets retry state, so reconnect() can call this without
   * clobbering its own retry counter.
   */
  private openConnection(): void {
    this.status.set(ConnectionStatus.CONNECTING);

    void this.client.start(this.url, this.options);

    if (this.autoCloseLifecycle) {
      this.scheduleAutoClose();
    }
  }

  private async reconnect() {
    if (this.isReconnecting) return;
    this.isReconnecting = true;

    try {
      // Keep retries in a single guarded task. Avoid self-recursion so
      // isReconnecting stays true for the entire retry cycle.
      while (true) {
        this.autoCloseLifecycle?.cancelScheduledPause();

        if (get(this.status) === ConnectionStatus.OPEN) return;

        const currentAttempts = get(this.retryAttempts);

        if (currentAttempts >= (this.params?.maxRetryAttempts ?? 0)) {
          this.status.set(ConnectionStatus.CLOSED);
          return;
        }

        if (currentAttempts > 0) {
          const delay = BACKOFF_DELAY * 2 ** currentAttempts;
          await asyncWait(delay);
        }

        this.retryAttempts.update((n) => n + 1);

        if (this.params?.onBeforeReconnect) {
          try {
            await this.params.onBeforeReconnect();
          } catch (err) {
            const errorArg = err instanceof Error ? err : new Error(String(err));
            this.events.emit("error", errorArg);
            // Treat hook failures like transport failures. The attempt already
            // counted, so continue in-loop and retry under the same guard.
            continue;
          }
        }

        this.openConnection();
        return;
      }
    } finally {
      this.isReconnecting = false;
    }
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

  // Fired by SSEFetchClient when the underlying fetch ends for any reason:
  //   1. Client-initiated pause (AbortError).
  //   2. Network error (transport failure).
  //   3. Application termination.
  // Client-initiated closes are ignored here; pause() handles its own
  // cleanup before setting status to PAUSED, so the guard below skips them.
  private handleCloseEvent = () => {
    const status = get(this.status);

    if (status !== ConnectionStatus.OPEN) return;

    // Server-initiated close. Reset retries if the connection had been
    // stable (open > MIN_STABLE_DURATION); unstable connections (e.g.
    // server opens then immediately closes) should still accumulate retries.
    // See also: pause() resets retries for client-initiated closes.
    const wasStable =
      this.openedAt !== null &&
      Date.now() - this.openedAt >= MIN_STABLE_DURATION;
    if (wasStable) {
      this.retryAttempts.set(0);
    }
    this.openedAt = null;

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
    this.openedAt = Date.now();
    this.status.set(ConnectionStatus.OPEN);

    // Once the connection has been stable for MIN_STABLE_DURATION, reset
    // retries so a future failure starts with a fresh budget. Mirrors the
    // wasStable check in handleCloseEvent for the case where the connection
    // errors out instead of closing cleanly. Especially important for
    // keep-alive consumers (e.g. cloud editor) that don't cycle through
    // pause(), which is the other place retries reset.
    setTimeout(() => {
      if (get(this.status) === ConnectionStatus.OPEN) {
        this.retryAttempts.set(0);
      }
    }, MIN_STABLE_DURATION);

    if (this.connectionCount > 1) {
      this.events.emit("reconnect");
    }
  };
}
