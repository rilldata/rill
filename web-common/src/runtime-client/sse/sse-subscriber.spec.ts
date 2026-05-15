import { describe, expect, it, vi } from "vitest";
import type { SSEConnection } from "./sse-connection";
import type { SSEMessage } from "./sse-protocol";
import { SSESubscriber } from "./sse-subscriber";

function fakeConnection() {
  const listeners = new Set<(msg: SSEMessage) => void>();
  const connection = {
    on: vi.fn((_event: string, listener: (msg: SSEMessage) => void) => {
      listeners.add(listener);
      return () => listeners.delete(listener);
    }),
  } as unknown as SSEConnection;

  return {
    connection,
    deliver(msg: SSEMessage) {
      listeners.forEach((l) => l(msg));
    },
  };
}

describe("SSESubscriber", () => {
  it("routes a tagged event through its decoder and emits the typed payload", () => {
    const { connection, deliver } = fakeConnection();
    const subscriber = new SSESubscriber<{ file: { path: string } }>(
      connection,
      { file: (data) => JSON.parse(data) as { path: string } },
    );

    const handler = vi.fn();
    subscriber.on("file", handler);

    deliver({ type: "file", data: JSON.stringify({ path: "/a.sql" }) });

    expect(handler).toHaveBeenCalledTimes(1);
    expect(handler).toHaveBeenCalledWith({ path: "/a.sql" });
  });

  it("normalizes an untagged frame to the 'message' decoder", () => {
    const { connection, deliver } = fakeConnection();
    const subscriber = new SSESubscriber<{ message: { text: string } }>(
      connection,
      { message: (data) => JSON.parse(data) as { text: string } },
    );

    const handler = vi.fn();
    subscriber.on("message", handler);

    deliver({ data: JSON.stringify({ text: "hi" }) });

    expect(handler).toHaveBeenCalledWith({ text: "hi" });
  });

  it("falls through to onUnknown if an untagged frame arrives with no 'message' decoder", () => {
    const { connection, deliver } = fakeConnection();
    const onUnknown = vi.fn();
    const subscriber = new SSESubscriber<{ file: { path: string } }>(
      connection,
      { file: (data) => JSON.parse(data) as { path: string } },
      { onUnknown },
    );

    const fileHandler = vi.fn();
    subscriber.on("file", fileHandler);

    const raw = { data: "untagged" };
    deliver(raw);

    expect(onUnknown).toHaveBeenCalledWith(raw);
    expect(fileHandler).not.toHaveBeenCalled();
  });

  it("calls onUnknown for an unregistered tagged event type", () => {
    const { connection, deliver } = fakeConnection();
    const onUnknown = vi.fn();
    new SSESubscriber<{ file: { path: string } }>(
      connection,
      { file: (data) => JSON.parse(data) as { path: string } },
      { onUnknown },
    );

    const raw = { type: "mystery", data: "x" };
    deliver(raw);

    expect(onUnknown).toHaveBeenCalledWith(raw);
  });

  it("calls onParseError when a decoder throws, and suppresses the typed emit", () => {
    const { connection, deliver } = fakeConnection();
    const onParseError = vi.fn();
    const subscriber = new SSESubscriber<{ file: { path: string } }>(
      connection,
      {
        file: () => {
          throw new Error("bad json");
        },
      },
      { onParseError },
    );

    const handler = vi.fn();
    subscriber.on("file", handler);

    const raw = { type: "file", data: "not json" };
    deliver(raw);

    expect(onParseError).toHaveBeenCalledTimes(1);
    const [err, message] = onParseError.mock.calls[0];
    expect((err as Error).message).toBe("bad json");
    expect(message).toBe(raw);
    expect(handler).not.toHaveBeenCalled();
  });
});
