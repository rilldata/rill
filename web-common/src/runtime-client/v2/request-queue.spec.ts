import { describe, it, expect, vi, beforeEach } from "vitest";
import { createQueueInterceptor, RequestQueue } from "./request-queue";

/**
 * Builds a minimal ConnectRPC-shaped request object.
 * Only the fields read by createQueueInterceptor are meaningful;
 * the rest are stubs to satisfy the interface.
 */
function fakeRequest(methodName: string, message: Record<string, unknown>) {
  return {
    stream: false as const,
    service: {},
    method: { name: methodName },
    url: "",
    init: {},
    signal: new AbortController().signal,
    header: new Headers(),
    contextValues: {},
    message,
  } as any; // eslint-disable-line @typescript-eslint/no-explicit-any
}

describe("createQueueInterceptor", () => {
  let sendRequest: (req: any) => Promise<any>; // eslint-disable-line @typescript-eslint/no-explicit-any
  let transport: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    const queue = new RequestQueue({ maxConcurrent: 10 });
    const interceptor = createQueueInterceptor(queue);

    // `transport` simulates the ConnectRPC transport layer (the `next` fn).
    // The interceptor wraps it, giving us `sendRequest`.
    transport = vi.fn().mockResolvedValue({ stream: false, message: {} });
    sendRequest = interceptor(transport as any); // eslint-disable-line @typescript-eslint/no-explicit-any
  });

  it("injects method-derived priority when message has no priority", async () => {
    const message: Record<string, unknown> = {
      metricsViewName: "test_view",
    };
    await sendRequest(fakeRequest("MetricsViewTimeRanges", message));

    expect(message.priority).toBe(100);
    expect(transport).toHaveBeenCalledOnce();
  });

  it("injects method-derived priority when message has priority 0", async () => {
    const message: Record<string, unknown> = {
      metricsViewName: "test_view",
      priority: 0,
    };
    await sendRequest(fakeRequest("MetricsViewAggregation", message));

    expect(message.priority).toBe(30);
  });

  it("preserves caller-set priority when non-zero", async () => {
    const message: Record<string, unknown> = {
      metricsViewName: "test_view",
      priority: 60,
    };
    await sendRequest(fakeRequest("MetricsViewAggregation", message));

    expect(message.priority).toBe(60);
  });

  it("falls back to DEFAULT_PRIORITY for unknown methods", async () => {
    const message: Record<string, unknown> = {};
    await sendRequest(fakeRequest("SomeUnknownMethod", message));

    expect(message.priority).toBe(10);
  });
});
