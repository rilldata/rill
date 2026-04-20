import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { SSEFetchClient, SSEHttpError } from "./sse-fetch-client";

function encodedStream(chunks: string[]): ReadableStream<Uint8Array> {
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

describe("SSEFetchClient", () => {
  const fetchSpy = vi.fn<typeof fetch>();

  beforeEach(() => {
    vi.stubGlobal("fetch", fetchSpy);
  });

  afterEach(() => {
    fetchSpy.mockReset();
    vi.unstubAllGlobals();
  });

  it("emits SSEHttpError with status and statusText on non-2xx", async () => {
    fetchSpy.mockResolvedValueOnce(
      new Response("nope", { status: 503, statusText: "Service Unavailable" }),
    );

    const client = new SSEFetchClient();
    const error = await new Promise<Error>((resolve) => {
      client.on("error", resolve);
      void client.start("http://x/sse");
    });

    expect(error).toBeInstanceOf(SSEHttpError);
    expect((error as SSEHttpError).status).toBe(503);
    expect((error as SSEHttpError).statusText).toBe("Service Unavailable");
  });

  it("emits a generic Error when the response has no body", async () => {
    const responseWithoutBody = {
      ok: true,
      status: 200,
      statusText: "OK",
      body: null,
    } as unknown as Response;
    fetchSpy.mockResolvedValueOnce(responseWithoutBody);

    const client = new SSEFetchClient();
    const error = await new Promise<Error>((resolve) => {
      client.on("error", resolve);
      void client.start("http://x/sse");
    });

    expect(error).toBeInstanceOf(Error);
    expect(error).not.toBeInstanceOf(SSEHttpError);
    expect(error.message).toContain("No response body");
  });

  it("does not emit an error when the stream is aborted", async () => {
    // Return a never-ending stream so the fetch's signal.abort() is the only
    // way the promise settles.
    fetchSpy.mockImplementationOnce((_url, init) => {
      return new Promise((_resolve, reject) => {
        init?.signal?.addEventListener("abort", () => {
          const err = new Error("aborted");
          err.name = "AbortError";
          reject(err);
        });
      });
    });

    const client = new SSEFetchClient();
    const errorHandler = vi.fn();
    client.on("error", errorHandler);

    const closed = new Promise<void>((resolve) => client.on("close", resolve));
    const started = client.start("http://x/sse");
    // Wait a microtask for the fetch to be in-flight before aborting.
    await Promise.resolve();
    client.stop();
    await started;
    await closed;

    expect(errorHandler).not.toHaveBeenCalled();
  });

  it("sends the JWT from getJwt as a Bearer Authorization header", async () => {
    fetchSpy.mockResolvedValueOnce(
      new Response(encodedStream(["data: ok\n\n"]), { status: 200 }),
    );

    const client = new SSEFetchClient();
    const closed = new Promise<void>((resolve) => client.on("close", resolve));
    void client.start("http://x/sse", { getJwt: () => "abc.def.ghi" });
    await closed;

    const [, init] = fetchSpy.mock.calls[0];
    expect((init as RequestInit).headers).toMatchObject({
      Authorization: "Bearer abc.def.ghi",
    });
  });

  it("stop() during streaming aborts the request and fires close", async () => {
    let abortSignal: AbortSignal | undefined;
    fetchSpy.mockImplementationOnce((_url, init) => {
      abortSignal = init?.signal ?? undefined;
      return new Promise((_resolve, reject) => {
        init?.signal?.addEventListener("abort", () => {
          const err = new Error("aborted");
          err.name = "AbortError";
          reject(err);
        });
      });
    });

    const client = new SSEFetchClient();
    const closed = new Promise<void>((resolve) => client.on("close", resolve));
    void client.start("http://x/sse");
    await Promise.resolve();

    client.stop();
    await closed;

    expect(abortSignal?.aborted).toBe(true);
  });

  it("cleanup() clears listeners so later emits are no-ops", async () => {
    fetchSpy.mockResolvedValueOnce(
      new Response(encodedStream(["data: ok\n\n"]), { status: 200 }),
    );

    const client = new SSEFetchClient();
    const messageHandler = vi.fn();
    client.on("message", messageHandler);

    const closed = new Promise<void>((resolve) => client.on("close", resolve));
    void client.start("http://x/sse");
    await closed;

    expect(messageHandler).toHaveBeenCalledTimes(1);

    client.cleanup();

    // Register a new listener and confirm cleanup dropped earlier subscribers
    // by firing another fetch and checking the original handler wasn't called
    // again.
    messageHandler.mockClear();

    fetchSpy.mockResolvedValueOnce(
      new Response(encodedStream(["data: second\n\n"]), { status: 200 }),
    );
    const closedAgain = new Promise<void>((resolve) =>
      client.on("close", resolve),
    );
    void client.start("http://x/sse");
    await closedAgain;

    expect(messageHandler).not.toHaveBeenCalled();
  });
});
