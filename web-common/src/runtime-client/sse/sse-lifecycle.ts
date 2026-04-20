import { Throttler } from "@rilldata/web-common/lib/throttler";
import type { SSEConnection } from "./sse-connection";

/**
 * Idle-timeout presets for SSELifecycle.
 *
 * - `aggressive`: pause the stream quickly when the tab idles. Used by Rill
 *   Developer where the browser's 6-connection per-host limit bites because
 *   SSE, queries, and dev assets all share `localhost:<port>`.
 * - `none`: don't attach a lifecycle at all. Used by the cloud editor and
 *   other consumers that need a persistent connection.
 */
export const LIFECYCLE_PRESETS = {
  aggressive: { short: 20_000, normal: 120_000 },
  none: undefined,
} as const satisfies Record<
  string,
  { short: number; normal: number } | undefined
>;

export type LifecyclePreset = keyof typeof LIFECYCLE_PRESETS;

/**
 * Abstracts the DOM-level signals that drive lifecycle decisions. Exists so
 * tests can drive the lifecycle without a real document/window.
 */
export interface LifecycleSignalSource {
  onVisibilityChange(listener: (visible: boolean) => void): () => void;
  onBlur(listener: () => void): () => void;
  onActivity(listener: () => void): () => void;
}

/**
 * Default signal source backed by the browser's document + window.
 */
export const domSignalSource: LifecycleSignalSource = {
  onVisibilityChange(listener) {
    const handler = () =>
      listener(document.visibilityState === "visible");
    document.addEventListener("visibilitychange", handler);
    return () => document.removeEventListener("visibilitychange", handler);
  },
  onBlur(listener) {
    const handler = () => listener();
    window.addEventListener("blur", handler);
    return () => window.removeEventListener("blur", handler);
  },
  onActivity(listener) {
    const handler = () => listener();
    window.addEventListener("click", handler);
    window.addEventListener("keydown", handler);
    window.addEventListener("focus", handler);
    return () => {
      window.removeEventListener("click", handler);
      window.removeEventListener("keydown", handler);
      window.removeEventListener("focus", handler);
    };
  },
};

export interface SSELifecycleOptions {
  /**
   * Override the signal source. Defaults to DOM events. Tests pass an
   * injected source to fire signals deterministically.
   */
  signals?: LifecycleSignalSource;
}

/**
 * Manages idle pausing for an SSEConnection based on tab visibility and
 * focus. Arms an idle throttler on blur/hide; cancels it and heartbeats
 * the connection on focus/show/activity.
 *
 * Optional by design. A consumer that wants a persistent connection simply
 * doesn't attach this layer.
 */
export class SSELifecycle {
  private readonly throttler: Throttler;
  private readonly unsubscribes: Array<() => void> = [];
  private started = false;

  constructor(
    private readonly connection: SSEConnection,
    private readonly idleTimeouts: { short: number; normal: number },
    private readonly options: SSELifecycleOptions = {},
  ) {
    this.throttler = new Throttler(idleTimeouts.normal, idleTimeouts.short);
  }

  /**
   * Attach DOM listeners and begin managing idle state. Idempotent; a
   * second call while already started is a no-op.
   */
  public start(): void {
    if (this.started) return;
    this.started = true;

    const signals = this.options.signals ?? domSignalSource;

    this.unsubscribes.push(
      signals.onVisibilityChange((visible) => {
        if (visible) {
          void this.connection.heartbeat();
        } else {
          this.scheduleIdlePause(true);
        }
      }),
      signals.onBlur(() => this.scheduleIdlePause(false)),
      signals.onActivity(() => void this.connection.heartbeat()),
    );
  }

  /**
   * Detach all listeners and cancel any pending idle pause. Safe to call
   * multiple times.
   */
  public stop(): void {
    this.throttler.cancel();
    while (this.unsubscribes.length) {
      const off = this.unsubscribes.pop();
      off?.();
    }
    this.started = false;
  }

  private scheduleIdlePause(useShortTimeout: boolean): void {
    this.throttler.cancel();
    this.throttler.throttle(() => this.connection.pause(), useShortTimeout);
  }
}
