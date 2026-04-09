import { describe, it, expect, vi, beforeEach } from "vitest";
import { createQueueInterceptor, RequestQueue } from "./request-queue";
import type { Interceptor } from "@connectrpc/connect";

/** Minimal mock of a ConnectRPC UnaryRequest for interceptor testing. */
function makeRequest(methodName: string, message: Record<string, unknown>) {
  return {
    stream: false as const,
    service: {} as any,
    method: { name: methodName } as any,
    url: "",
    init: {},
    signal: new AbortController().signal,
    header: new Headers(),
    contextValues: {} as any,
    message: message as any,
  };
}

describe("createQueueInterceptor", () => {
  let interceptor: Interceptor;
  let next: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    const queue = new RequestQueue({ maxConcurrent: 10 });
    interceptor = createQueueInterceptor(queue);
    next = vi.fn().mockResolvedValue({ stream: false, message: {} });
  });

  it("injects method-derived priority when message has no priority", async () => {
    const message: Record<string, unknown> = {
      metricsViewName: "test_view",
    };
    await interceptor(next)(makeRequest("MetricsViewTimeRanges", message));

    expect(message.priority).toBe(100);
    expect(next).toHaveBeenCalledOnce();
  });

  it("injects method-derived priority when message has priority 0", async () => {
    const message: Record<string, unknown> = {
      metricsViewName: "test_view",
      priority: 0,
    };
    await interceptor(next)(makeRequest("MetricsViewAggregation", message));

    expect(message.priority).toBe(30);
  });

  it("preserves caller-set priority when non-zero", async () => {
    const message: Record<string, unknown> = {
      metricsViewName: "test_view",
      priority: 60,
    };
    await interceptor(next)(makeRequest("MetricsViewAggregation", message));

    expect(message.priority).toBe(60);
  });

  it("falls back to DEFAULT_PRIORITY for unknown methods", async () => {
    const message: Record<string, unknown> = {};
    await interceptor(next)(makeRequest("SomeUnknownMethod", message));

    expect(message.priority).toBe(10);
  });
});
