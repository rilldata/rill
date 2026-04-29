import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  SSEConnectionLifecycle,
  type LifecycleSignalSource,
  type LifecycleControl,
} from "./sse-connection-lifecycle";

function fakeConnection() {
  return {
    pause: vi.fn(),
    resumeIfPaused: vi.fn().mockResolvedValue(undefined),
  } as unknown as LifecycleControl & {
    pause: ReturnType<typeof vi.fn>;
    resumeIfPaused: ReturnType<typeof vi.fn>;
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

describe("SSEConnectionLifecycle", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("arms an idle pause when the tab becomes hidden", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSEConnectionLifecycle(
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

  it("resumes when the tab becomes visible again and cancels pending idle pause", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSEConnectionLifecycle(
      conn,
      { short: 100, normal: 1_000 },
      { signals: signals.signals },
    );
    lifecycle.start();

    signals.fireVisibility(false);
    vi.advanceTimersByTime(50);
    signals.fireVisibility(true);

    vi.advanceTimersByTime(1_000);
    expect(conn.pause).not.toHaveBeenCalled();
    expect(conn.resumeIfPaused).toHaveBeenCalledTimes(1);
  });

  it("arms an idle pause on blur (long timeout)", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSEConnectionLifecycle(
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

  it("resumes on activity and cancels pending blur pause", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSEConnectionLifecycle(
      conn,
      { short: 100, normal: 1_000 },
      { signals: signals.signals },
    );
    lifecycle.start();

    signals.fireBlur();
    vi.advanceTimersByTime(500);
    signals.fireActivity();

    vi.advanceTimersByTime(1_000);
    expect(conn.pause).not.toHaveBeenCalled();
    expect(conn.resumeIfPaused).toHaveBeenCalledTimes(1);
  });

  it("stop() detaches listeners and cancels the pending idle pause", () => {
    const conn = fakeConnection();
    const signals = fakeSignals();
    const lifecycle = new SSEConnectionLifecycle(
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
    const lifecycle = new SSEConnectionLifecycle(
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
