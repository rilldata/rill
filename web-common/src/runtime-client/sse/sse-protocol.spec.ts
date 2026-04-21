import { describe, expect, it } from "vitest";
import {
  hasDispatchableData,
  isEventBoundary,
  parseSSELine,
  readSSEStream,
  type SSEMessage,
} from "./sse-protocol";

function streamFrom(chunks: string[]): ReadableStream<Uint8Array> {
  const encoder = new TextEncoder();
  return new ReadableStream({
    start(controller) {
      for (const chunk of chunks) {
        controller.enqueue(encoder.encode(chunk));
      }
      controller.close();
    },
  });
}

async function collect(
  stream: ReadableStream<Uint8Array>,
): Promise<SSEMessage[]> {
  const out: SSEMessage[] = [];
  for await (const msg of readSSEStream(stream)) {
    out.push(msg);
  }
  return out;
}

describe("readSSEStream", () => {
  it("parses a single event with a single data line", async () => {
    const messages = await collect(streamFrom(["data: hello\n\n"]));
    expect(messages).toEqual([{ data: "hello" }]);
  });

  it("accumulates multi-line data fields with newlines", async () => {
    const messages = await collect(
      streamFrom(["data: one\ndata: two\ndata: three\n\n"]),
    );
    expect(messages).toEqual([{ data: "one\ntwo\nthree" }]);
  });

  it("honors `event:` to set message type", async () => {
    const messages = await collect(streamFrom(["event: file\ndata: {}\n\n"]));
    expect(messages).toEqual([{ type: "file", data: "{}" }]);
  });

  it("ignores comment lines", async () => {
    const messages = await collect(streamFrom([":keepalive\n\ndata: hi\n\n"]));
    expect(messages).toEqual([{ data: "hi" }]);
  });

  it("ignores backend keepalive comments (`:ok` and `:keepalive`)", async () => {
    const messages = await collect(
      streamFrom([":ok\n\n:keepalive\n\ndata: hi\n\n:keepalive\n\n"]),
    );
    expect(messages).toEqual([{ data: "hi" }]);
  });

  it("treats the empty line as an event boundary", async () => {
    const messages = await collect(streamFrom(["data: a\n\ndata: b\n\n"]));
    expect(messages).toEqual([{ data: "a" }, { data: "b" }]);
  });

  it("reassembles an event split across two chunks", async () => {
    const messages = await collect(
      streamFrom(["event: file\nda", "ta: {}\n\n"]),
    );
    expect(messages).toEqual([{ type: "file", data: "{}" }]);
  });

  it("handles CRLF line endings", async () => {
    const messages = await collect(
      streamFrom(["event: file\r\ndata: {}\r\n\r\n"]),
    );
    expect(messages).toEqual([{ type: "file", data: "{}" }]);
  });

  it("holds a trailing partial line until the next chunk", async () => {
    const messages = await collect(streamFrom(["data: par", "tial\n\n"]));
    expect(messages).toEqual([{ data: "partial" }]);
  });

  it("ignores unknown fields like `id:` and `retry:`", async () => {
    const messages = await collect(
      streamFrom(["id: 42\nretry: 3000\ndata: hi\n\n"]),
    );
    expect(messages).toEqual([{ data: "hi" }]);
  });

  it("yields a final event when the stream ends without a trailing newline or boundary", async () => {
    const messages = await collect(streamFrom(["data: hi"]));
    expect(messages).toEqual([{ data: "hi" }]);
  });

  it("does not yield data-less events", async () => {
    const messages = await collect(streamFrom([":comment\n\n\n"]));
    expect(messages).toEqual([]);
  });
});

describe("parseSSELine", () => {
  it("skips empty lines", () => {
    const event: Partial<SSEMessage> = {};
    parseSSELine("", event);
    expect(event).toEqual({});
  });

  it("skips comment lines", () => {
    const event: Partial<SSEMessage> = {};
    parseSSELine(":keepalive", event);
    expect(event).toEqual({});
  });

  it("parses event type", () => {
    const event: Partial<SSEMessage> = {};
    parseSSELine("event: file", event);
    expect(event.type).toBe("file");
  });

  it("parses data", () => {
    const event: Partial<SSEMessage> = {};
    parseSSELine("data: hello", event);
    expect(event.data).toBe("hello");
  });

  it("strips only one leading space after ':'", () => {
    const event: Partial<SSEMessage> = {};
    parseSSELine("data:  hello", event);
    expect(event.data).toBe(" hello");
  });

  it("preserves trailing spaces in data", () => {
    const event: Partial<SSEMessage> = {};
    parseSSELine("data: hello  ", event);
    expect(event.data).toBe("hello  ");
  });

  it("accumulates multi-line data with newlines", () => {
    const event: Partial<SSEMessage> = {};
    parseSSELine("data: one", event);
    parseSSELine("data: two", event);
    expect(event.data).toBe("one\ntwo");
  });

  it("ignores unknown fields (id, retry)", () => {
    const event: Partial<SSEMessage> = {};
    parseSSELine("id: 123", event);
    parseSSELine("retry: 5000", event);
    expect(event).toEqual({});
  });
});

describe("isEventBoundary", () => {
  it("treats empty line as event boundary", () => {
    expect(isEventBoundary("")).toBe(true);
  });

  it("treats non-empty line as in-progress", () => {
    expect(isEventBoundary("data: x")).toBe(false);
    expect(isEventBoundary("   ")).toBe(false);
  });
});

describe("hasDispatchableData", () => {
  it("requires non-empty data", () => {
    expect(hasDispatchableData({})).toBe(false);
    expect(hasDispatchableData({ data: "" })).toBe(false);
    expect(hasDispatchableData({ data: "x" })).toBe(true);
  });
});
