import { describe, it, expect, vi, beforeEach } from "vitest";
import { createQueueInterceptor, RequestQueue } from "./request-queue";

/** Minimal mock of a ConnectRPC UnaryRequest for interceptor testing. */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function makeRequest(
  methodName: string,
  message: Record<string, unknown>,
): any {
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
  };
}

describe("createQueueInterceptor", () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  let run: (req: any) => Promise<any>;
  let next: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    const queue = new RequestQueue({ maxConcurrent: 10 });
    const interceptor = createQueueInterceptor(queue);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    next = vi.fn().mockResolvedValue({ stream: false, message: {} }) as any;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    run = interceptor(next as any);
  });

  it("injects method-derived priority when message has no priority", async () => {
    const message: Record<string, unknown> = {
      metricsViewName: "test_view",
    };
    await run(makeRequest("MetricsViewTimeRanges", message));

    expect(message.priority).toBe(100);
    expect(next).toHaveBeenCalledOnce();
  });

  it("injects method-derived priority when message has priority 0", async () => {
    const message: Record<string, unknown> = {
      metricsViewName: "test_view",
      priority: 0,
    };
    await run(makeRequest("MetricsViewAggregation", message));

    expect(message.priority).toBe(30);
  });

  it("preserves caller-set priority when non-zero", async () => {
    const message: Record<string, unknown> = {
      metricsViewName: "test_view",
      priority: 60,
    };
    await run(makeRequest("MetricsViewAggregation", message));

    expect(message.priority).toBe(60);
  });

  it("falls back to DEFAULT_PRIORITY for unknown methods", async () => {
    const message: Record<string, unknown> = {};
    await run(makeRequest("SomeUnknownMethod", message));

    expect(message.priority).toBe(10);
  });
});
