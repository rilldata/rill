import { writable } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";

class FakeConnection {
  public status = writable("closed");
  private readonly handlers = new Map<string, Set<(arg?: unknown) => void>>();

  public start = vi.fn();
  public pause = vi.fn();
  public resumeIfPaused = vi.fn(async () => {});
  public close = vi.fn();
  public on = (event: string, listener: (arg?: unknown) => void) => {
    if (!this.handlers.has(event)) this.handlers.set(event, new Set());
    this.handlers.get(event)!.add(listener);
    return () => this.handlers.get(event)!.delete(listener);
  };
  public once = this.on;

  public fire(event: string, payload?: unknown) {
    this.handlers.get(event)?.forEach((handler) => handler(payload));
  }
}

class FakeSubscriber {
  private readonly handlers = new Map<string, Set<(arg?: unknown) => void>>();
  public cleanup = vi.fn();

  public on = (event: string, listener: (arg?: unknown) => void) => {
    if (!this.handlers.has(event)) this.handlers.set(event, new Set());
    this.handlers.get(event)!.add(listener);
    return () => this.handlers.get(event)!.delete(listener);
  };
  public once = this.on;

  public fire(event: string, payload?: unknown) {
    this.handlers.get(event)?.forEach((handler) => handler(payload));
  }
}

class FakeLifecycle {
  public start = vi.fn();
  public stop = vi.fn();
}

const fakeConnections: FakeConnection[] = [];
const fakeSubscribers: FakeSubscriber[] = [];
const fakeLifecycles: FakeLifecycle[] = [];

vi.mock("./sse-connection", () => {
  return {
    SSEConnection: class {
      constructor() {
        const conn = new FakeConnection();
        fakeConnections.push(conn);
        return conn;
      }
    },
  };
});

vi.mock("./sse-subscriber", () => {
  return {
    SSESubscriber: class {
      constructor() {
        const subscriber = new FakeSubscriber();
        fakeSubscribers.push(subscriber);
        return subscriber;
      }
    },
  };
});

vi.mock("./sse-connection-lifecycle", () => {
  return {
    SSEConnectionLifecycle: class {
      constructor() {
        const lifecycle = new FakeLifecycle();
        fakeLifecycles.push(lifecycle);
        return lifecycle;
      }
    },
  };
});

import { createSSEStream } from "./create-sse-stream";

function latestConnection() {
  return fakeConnections[fakeConnections.length - 1];
}

function latestSubscriber() {
  return fakeSubscribers[fakeSubscribers.length - 1];
}

function latestLifecycle() {
  return fakeLifecycles[fakeLifecycles.length - 1];
}

describe("createSSEStream", () => {
  beforeEach(() => {
    fakeConnections.length = 0;
    fakeSubscribers.length = 0;
    fakeLifecycles.length = 0;
    vi.clearAllMocks();
  });

  it("starts the underlying connection and lifecycle", () => {
    const stream = createSSEStream<{ message: string }>({
      connection: { maxRetryAttempts: 3 },
      decoders: { message: (data) => data },
      lifecycle: { idleTimeouts: { short: 10, normal: 20 } },
    });

    stream.start("http://x/sse", { method: "POST", body: { prompt: "hi" } });

    expect(latestConnection().start).toHaveBeenCalledWith("http://x/sse", {
      method: "POST",
      body: { prompt: "hi" },
    });
    expect(latestLifecycle().start).toHaveBeenCalledTimes(1);
  });

  it("forwards typed and connection events", () => {
    const stream = createSSEStream<{ message: string }>({
      decoders: { message: (data) => data },
    });
    const typedHandler = vi.fn();
    const connectionErrorHandler = vi.fn();

    stream.on("message", typedHandler);
    stream.onConnection("error", connectionErrorHandler);

    latestSubscriber().fire("message", "hello");
    const err = new Error("boom");
    latestConnection().fire("error", err);

    expect(typedHandler).toHaveBeenCalledWith("hello");
    expect(connectionErrorHandler).toHaveBeenCalledWith(err);
  });

  it("close(true) tears down subscriber and lifecycle", () => {
    const stream = createSSEStream<{ message: string }>({
      decoders: { message: (data) => data },
      lifecycle: { idleTimeouts: { short: 10, normal: 20 } },
    });

    stream.close(true);

    expect(latestConnection().close).toHaveBeenCalledWith(true);
    expect(latestSubscriber().cleanup).toHaveBeenCalledTimes(1);
    expect(latestLifecycle().stop).toHaveBeenCalledTimes(1);
  });

  it("cleanup always closes with cleanup=true", () => {
    const stream = createSSEStream<{ message: string }>({
      decoders: { message: (data) => data },
      lifecycle: { idleTimeouts: { short: 10, normal: 20 } },
    });

    stream.cleanup();

    expect(latestConnection().close).toHaveBeenCalledWith(true);
    expect(latestSubscriber().cleanup).toHaveBeenCalledTimes(1);
    expect(latestLifecycle().stop).toHaveBeenCalledTimes(1);
  });
});
