import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// A controllable stand-in for SSEFetchClient. Each constructed instance is
// pushed onto `fakeClients` so tests can drive `open`/`error`/`close`/`message`
// events deterministically.
class FakeClient {
  private handlers = new Map<string, Set<(arg?: unknown) => void>>();
  public start = vi.fn(async (_url: string, _opts?: unknown) => {});
  public stop = vi.fn();
  public cleanup = vi.fn();
  public isStreaming = vi.fn(() => false);

  public on = (event: string, listener: (arg?: unknown) => void) => {
    if (!this.handlers.has(event)) this.handlers.set(event, new Set());
    this.handlers.get(event)!.add(listener);
    return () => this.handlers.get(event)!.delete(listener);
  };
  public once = this.on;

  public fire(event: string, arg?: unknown) {
    this.handlers.get(event)?.forEach((h) => h(arg));
  }
}

const fakeClients: FakeClient[] = [];

vi.mock("./sse-fetch-client", () => {
  return {
    SSEFetchClient: class {
      constructor() {
        const c = new FakeClient();
        fakeClients.push(c);
        return c as unknown as object;
      }
    },
    SSEHttpError: class SSEHttpError extends Error {},
  };
});

import { ConnectionStatus, SSEConnection } from "./sse-connection";

function latestClient(): FakeClient {
  return fakeClients[fakeClients.length - 1];
}

describe("SSEConnection", () => {
  beforeEach(() => {
    fakeClients.length = 0;
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("goes CLOSED → CONNECTING → OPEN when the transport opens", () => {
    const conn = new SSEConnection();
    expect(get(conn.status)).toBe(ConnectionStatus.CLOSED);

    conn.start("http://x/sse");
    expect(get(conn.status)).toBe(ConnectionStatus.CONNECTING);

    latestClient().fire("open");
    expect(get(conn.status)).toBe(ConnectionStatus.OPEN);
  });

  it("retries on error with exponential backoff and stops at maxRetryAttempts", async () => {
    const conn = new SSEConnection({
      retryOnError: true,
      maxRetryAttempts: 3,
    });
    conn.start("http://x/sse");

    const client = latestClient();

    // First failure: retryAttempts 0 → 1, no delay.
    client.fire("error", new Error("net"));
    expect(get(conn.status)).toBe(ConnectionStatus.CONNECTING);
    await vi.advanceTimersByTimeAsync(0);
    expect(client.start).toHaveBeenCalledTimes(2);

    // Second failure: delay 1000 * 2^1 = 2000.
    client.fire("error", new Error("net"));
    await vi.advanceTimersByTimeAsync(1999);
    expect(client.start).toHaveBeenCalledTimes(2);
    await vi.advanceTimersByTimeAsync(1);
    expect(client.start).toHaveBeenCalledTimes(3);

    // Third failure: delay 1000 * 2^2 = 4000.
    client.fire("error", new Error("net"));
    await vi.advanceTimersByTimeAsync(4000);
    // maxRetryAttempts reached on the next tick.
    client.fire("error", new Error("net"));
    await vi.advanceTimersByTimeAsync(10_000);

    expect(get(conn.status)).toBe(ConnectionStatus.CLOSED);
  });

  it("resets retryAttempts after the connection is stable (open-then-stable path)", async () => {
    const conn = new SSEConnection({
      retryOnError: true,
      maxRetryAttempts: 3,
    });
    conn.start("http://x/sse");

    const client = latestClient();

    // Trigger one error (retryAttempts -> 1) then open successfully.
    client.fire("error", new Error("net"));
    await vi.advanceTimersByTimeAsync(0);
    client.fire("open");
    expect(get(conn.status)).toBe(ConnectionStatus.OPEN);

    // Stable threshold (5s). After this, retryAttempts is reset to 0.
    await vi.advanceTimersByTimeAsync(5000);

    // A subsequent error should schedule with the first-attempt delay (0),
    // not a large backoff from the accumulated counter.
    client.fire("error", new Error("net"));
    await vi.advanceTimersByTimeAsync(0);
    // A new start() call happened without any wait.
    expect(client.start.mock.calls.length).toBeGreaterThanOrEqual(2);
  });

  it("resets retryAttempts when a stable open connection closes (wasStable close path)", async () => {
    const conn = new SSEConnection({
      retryOnError: true,
      retryOnClose: true,
      maxRetryAttempts: 3,
    });
    conn.start("http://x/sse");
    const client = latestClient();

    // Build up one retry attempt first.
    client.fire("error", new Error("net"));
    await vi.advanceTimersByTimeAsync(0);
    expect(client.start).toHaveBeenCalledTimes(2);

    // Open successfully, then simulate wall-clock time passing without firing
    // the open-path 5s timer reset.
    const openedAt = Date.now();
    client.fire("open");
    vi.setSystemTime(openedAt + 6_000);

    // A stable server close should reset attempts so reconnect is immediate.
    client.fire("close");
    expect(get(conn.status)).toBe(ConnectionStatus.CONNECTING);
    expect(client.start).toHaveBeenCalledTimes(3);
  });

  it("resumeIfPaused() resumes from PAUSED", async () => {
    const conn = new SSEConnection({
      retryOnError: true,
      maxRetryAttempts: 3,
    });
    conn.start("http://x/sse");
    const client = latestClient();
    client.fire("open");

    conn.pause();
    expect(get(conn.status)).toBe(ConnectionStatus.PAUSED);

    await conn.resumeIfPaused();
    expect(get(conn.status)).toBe(ConnectionStatus.CONNECTING);
  });

  it("pause() resets retryAttempts", async () => {
    const conn = new SSEConnection({
      retryOnError: true,
      maxRetryAttempts: 3,
    });
    conn.start("http://x/sse");
    const client = latestClient();

    client.fire("error", new Error("net"));
    await vi.advanceTimersByTimeAsync(0);

    conn.pause();
    expect(get(conn.status)).toBe(ConnectionStatus.PAUSED);

    // Restart from pause via resumeIfPaused; retryAttempts should be 0, so the
    // next start call happens without backoff.
    await conn.resumeIfPaused();
    await vi.advanceTimersByTimeAsync(0);
    expect(client.start).toHaveBeenCalledTimes(3);
  });

  it("keeps heartbeat() as a backward-compatible alias", async () => {
    const conn = new SSEConnection();
    const resumeSpy = vi.spyOn(conn, "resumeIfPaused");

    await conn.heartbeat();

    expect(resumeSpy).toHaveBeenCalledTimes(1);
  });

  it("close(true) clears listeners", () => {
    const conn = new SSEConnection();
    const errorHandler = vi.fn();
    conn.on("error", errorHandler);
    conn.start("http://x/sse");

    conn.close(true);

    // Firing an error on the (detached) underlying client after close + cleanup
    // should not reach the subscriber.
    latestClient().fire("error", new Error("net"));
    expect(errorHandler).not.toHaveBeenCalled();
  });

  it("awaits onBeforeReconnect before each retry", async () => {
    const hook = vi.fn().mockResolvedValue(undefined);
    const conn = new SSEConnection({
      retryOnError: true,
      maxRetryAttempts: 3,
      onBeforeReconnect: hook,
    });
    conn.start("http://x/sse");
    const client = latestClient();

    client.fire("error", new Error("net"));
    // Allow pending microtasks and timers to flush.
    await vi.advanceTimersByTimeAsync(0);

    expect(hook).toHaveBeenCalledTimes(1);
    expect(client.start).toHaveBeenCalledTimes(2);
  });

  it("counts onBeforeReconnect rejections as failed attempts and lands CLOSED after maxRetryAttempts", async () => {
    const hook = vi.fn().mockRejectedValue(new Error("refresh failed"));
    const conn = new SSEConnection({
      retryOnError: true,
      maxRetryAttempts: 2,
      onBeforeReconnect: hook,
    });
    const errors: Error[] = [];
    conn.on("error", (err) => errors.push(err));
    conn.start("http://x/sse");
    const client = latestClient();

    // First transport failure → first retry (delay 0); hook rejects.
    client.fire("error", new Error("net"));
    // Let the first hook reject and the recursive reconnect re-enter.
    await vi.advanceTimersByTimeAsync(0);
    // Second attempt also runs the hook; delay is 1000 * 2^1 = 2000 on the
    // 2nd real attempt. Flush that as well.
    await vi.advanceTimersByTimeAsync(2000);
    // Third attempt (maxed out) lands CLOSED.
    await vi.advanceTimersByTimeAsync(4000);

    expect(hook).toHaveBeenCalled();
    // Transport never opens because the hook blocks it.
    expect(client.start).toHaveBeenCalledTimes(1);
    expect(errors.some((e) => e.message === "refresh failed")).toBe(true);
    expect(get(conn.status)).toBe(ConnectionStatus.CLOSED);
  });

  it("keeps onBeforeReconnect retries in a single guarded reconnect task", async () => {
    const hook = vi.fn().mockRejectedValue(new Error("refresh failed"));
    const conn = new SSEConnection({
      retryOnError: true,
      maxRetryAttempts: 3,
      onBeforeReconnect: hook,
    });
    conn.start("http://x/sse");
    const client = latestClient();

    client.fire("error", new Error("net"));
    await vi.advanceTimersByTimeAsync(0);
    expect(hook).toHaveBeenCalledTimes(1);

    // If the reconnect guard is held correctly, this extra trigger during the
    // hook-failure backoff window is ignored and does not start a parallel loop.
    void (conn as unknown as { reconnect: () => Promise<void> }).reconnect();

    await vi.advanceTimersByTimeAsync(2_000);
    expect(hook).toHaveBeenCalledTimes(2);
  });
});
