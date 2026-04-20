import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { SSELifecycle, type LifecycleSignalSource } from "./sse-lifecycle";
import type { SSEConnection } from "./sse-connection";

function fakeConnection() {
  return {
    pause: vi.fn(),
    heartbeat: vi.fn().mockResolvedValue(undefined),
  } as unknown as SSEConnection & {
    pause: ReturnType<typeof vi.fn>;
    heartbeat: ReturnType<typeof vi.fn>;
  };
}

function fakeSignals() {
  const visibility = new Set<(visible: boolean) => void>();
  const blur = new Set<() => void>();
  const activity = new Set<() => void>();

  const signals: LifecycleSignalSource = {
    onVisibilityChange(listener) {
      visibility.add(listener);
      return () => visibility.delete(listener);
    },
    onBlur(listener) {
      blur.add(listener);
      return () => blur.delete(listener);
    },
    onActivity(listener) {
      activity.add(listener);
      return () => activity.delete(listener);
    },
  };

  return {
    signals,
    fireVisibility(visible: boolean) {
      visibility.forEach((l) => l(visible));
    },
    fireBlur() {
      blur.forEach((l) => l());
    },
    fireActivity() {
      activity.forEach((l) => l());
    },
    listenerCount() {
      return visibility.size + blur.size + activity.size;
    },
  };
}

describe("SSELifecycle", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("arms an idle pause when the tab becomes hidden", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSELifecycle(
      conn,
      { short: 100, normal: 1_000 },
      { signals: signals.signals },
    );
    lifecycle.start();

    signals.fireVisibility(false);
    // Hidden uses the short timeout.
    expect(conn.pause).not.toHaveBeenCalled();
    vi.advanceTimersByTime(99);
    expect(conn.pause).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1);
    expect(conn.pause).toHaveBeenCalledTimes(1);
  });

  it("heartbeats when the tab becomes visible again", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSELifecycle(
      conn,
      { short: 100, normal: 1_000 },
      { signals: signals.signals },
    );
    lifecycle.start();

    signals.fireVisibility(true);
    expect(conn.heartbeat).toHaveBeenCalledTimes(1);
  });

  it("arms an idle pause on blur (long timeout)", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSELifecycle(
      conn,
      { short: 100, normal: 1_000 },
      { signals: signals.signals },
    );
    lifecycle.start();

    signals.fireBlur();
    vi.advanceTimersByTime(999);
    expect(conn.pause).not.toHaveBeenCalled();
    vi.advanceTimersByTime(1);
    expect(conn.pause).toHaveBeenCalledTimes(1);
  });

  it("heartbeats on activity", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSELifecycle(
      conn,
      { short: 100, normal: 1_000 },
      { signals: signals.signals },
    );
    lifecycle.start();

    signals.fireActivity();
    expect(conn.heartbeat).toHaveBeenCalledTimes(1);
  });

  it("stop() detaches listeners and cancels the pending idle pause", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSELifecycle(
      conn,
      { short: 100, normal: 1_000 },
      { signals: signals.signals },
    );
    lifecycle.start();
    expect(signals.listenerCount()).toBeGreaterThan(0);

    signals.fireBlur();
    lifecycle.stop();

    vi.advanceTimersByTime(5_000);
    expect(conn.pause).not.toHaveBeenCalled();
    expect(signals.listenerCount()).toBe(0);
  });

  it("start() is idempotent", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSELifecycle(
      conn,
      { short: 100, normal: 1_000 },
      { signals: signals.signals },
    );
    lifecycle.start();
    const first = signals.listenerCount();
    lifecycle.start();
    expect(signals.listenerCount()).toBe(first);
  });
});
