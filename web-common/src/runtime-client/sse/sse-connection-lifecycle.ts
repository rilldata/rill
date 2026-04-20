import { Throttler } from "@rilldata/web-common/lib/throttler";

/**
 * Narrow connection contract needed by SSEConnectionLifecycle.
 * Keeps lifecycle policy decoupled from concrete connection implementations.
 */
export interface LifecycleControl {
  pause(): void;
  resumeIfPaused(): Promise<void>;
}

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
    const handler = () => listener(document.visibilityState === "visible");
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

export interface SSEConnectionLifecycleOptions {
  /**
   * Override the signal source. Defaults to DOM events. Tests pass an
   * injected source to fire signals deterministically.
   */
  signals?: LifecycleSignalSource;
}

/**
 * Manages idle pausing for an SSE stream connection based on tab visibility
 * and focus. Schedules a delayed pause on blur/hide, and cancels that pending
 * pause while resuming on focus/show/activity.
 *
 * Optional by design. A consumer that wants a persistent connection simply
 * doesn't attach this layer.
 */
export class SSEConnectionLifecycle {
  private readonly throttler: Throttler;
  private readonly unsubscribes: Array<() => void> = [];
  private started = false;

  constructor(
    private readonly connection: LifecycleControl,
    idleTimeouts: { short: number; normal: number },
    private readonly options: SSEConnectionLifecycleOptions = {},
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
          this.resumeAndCancelPendingPause();
        } else {
          this.schedulePauseWithTimeout(true);
        }
      }),
      signals.onBlur(() => this.schedulePauseWithTimeout(false)),
      signals.onActivity(() => this.resumeAndCancelPendingPause()),
    );
  }

  /**
   * Detach all listeners and cancel any pending idle pause. Safe to call
   * multiple times.
   */
  public stop(): void {
    this.cancelScheduledPause();
    while (this.unsubscribes.length) {
      const off = this.unsubscribes.pop();
      off?.();
    }
    this.started = false;
  }

  /**
   * Schedule a pause with either the short (prioritized) or normal timeout.
   * Public for backwards-compatible callers that still drive auto-close
   * manually through deprecated SSEConnection methods.
   */
  public schedulePause(prioritize: boolean): void {
    this.schedulePauseWithTimeout(prioritize);
  }

  /**
   * Cancel any pending pause.
   */
  public cancelScheduledPause(): void {
    this.throttler.cancel();
  }

  private schedulePauseWithTimeout(useShortTimeout: boolean): void {
    this.throttler.cancel();
    this.throttler.throttle(() => this.connection.pause(), useShortTimeout);
  }

  private resumeAndCancelPendingPause(): void {
    this.cancelScheduledPause();
    void this.connection.resumeIfPaused();
  }
}
