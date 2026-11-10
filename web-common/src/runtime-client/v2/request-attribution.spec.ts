import {
  Code,
  ConnectError,
  createClient,
  type UnaryRequest,
  type UnaryResponse,
} from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { describe, expect, it } from "vitest";
import { QueryService } from "../../proto/gen/rill/runtime/v1/queries_connect";
import {
  createRequestAttributionInterceptor,
  getAttributedRequests,
} from "./request-attribution";

function request(
  methodName: string,
  signal: AbortSignal,
  message: Record<string, unknown> = {},
  serviceTypeName = "rill.runtime.v1.QueryService",
): UnaryRequest {
  return {
    service: { typeName: serviceTypeName },
    method: { name: methodName },
    message,
    signal,
  } as unknown as UnaryRequest;
}

const pending = () => new Promise<UnaryResponse>(() => {});

// In browsers the default abort reason is a DOMException that extends Error; jsdom's does not, so build it as browsers see it
function abortError(message: string): Error {
  const error = new Error(message);
  error.name = "AbortError";
  return error;
}

/** A QueryService client on a real Connect transport whose fetch is `fetchImpl`. */
function queryServiceClient(fetchImpl: typeof fetch) {
  return createClient(
    QueryService,
    createConnectTransport({
      baseUrl: "http://localhost:9009",
      fetch: fetchImpl,
      interceptors: [createRequestAttributionInterceptor()],
    }),
  );
}

// Like a browser fetch: rejects with the signal's reason once it aborts
const fetchUntilAborted: typeof fetch = (
  _input: RequestInfo | URL,
  init?: RequestInit,
) =>
  new Promise<Response>((_, reject) => {
    const signal = init?.signal;
    if (signal?.aborted) reject(signal.reason);
    signal?.addEventListener("abort", () => reject(signal.reason));
  });

describe("createRequestAttributionInterceptor", () => {
  it("tags the abort reason of a request cancelled through its signal", () => {
    const controller = new AbortController();
    void createRequestAttributionInterceptor()(pending)(
      request("MetricsViewAggregation", controller.signal, {
        metricsView: "revenue",
      }),
    );
    const reason = new Error("signal is aborted without reason");

    controller.abort(reason);

    expect(getAttributedRequests(reason)).toEqual([
      { api: "QueryService/metricsViewAggregation", resource: "revenue" },
    ]);
  });

  it("tags the error a request fails with, and rethrows it as the ConnectError the transport would", async () => {
    const error = new TypeError("Failed to fetch");

    const thrown = await createRequestAttributionInterceptor()(() =>
      Promise.reject(error),
    )(
      request(
        "GetResource",
        new AbortController().signal,
        {},
        "rill.runtime.v1.RuntimeService",
      ),
    ).catch((e: unknown) => e);

    expect(thrown).toBeInstanceOf(ConnectError);
    expect((thrown as ConnectError).cause).toBe(error);
    expect(getAttributedRequests(error)).toEqual([
      { api: "RuntimeService/getResource" },
    ]);
    expect(getAttributedRequests(thrown)).toEqual([
      { api: "RuntimeService/getResource" },
    ]);
  });

  it("rethrows a ConnectError as it is", async () => {
    const error = new ConnectError("jwt is expired", Code.Unauthenticated);

    await expect(
      createRequestAttributionInterceptor()(() => Promise.reject(error))(
        request("MetricsViewAggregation", new AbortController().signal),
      ),
    ).rejects.toBe(error);

    expect(getAttributedRequests(error)).toEqual([
      { api: "QueryService/metricsViewAggregation" },
    ]);
  });

  it("tags a request only once when its abort also rejects it", async () => {
    const controller = new AbortController();
    const rejectOnAbort = () =>
      new Promise<UnaryResponse>((_, reject) =>
        controller.signal.addEventListener("abort", () =>
          reject(controller.signal.reason),
        ),
      );
    const call = createRequestAttributionInterceptor()(rejectOnAbort)(
      request("MetricsViewAggregation", controller.signal),
    );
    const reason = abortError("signal is aborted without reason");

    controller.abort(reason);
    const thrown = await call.catch((e: unknown) => e);

    expect(getAttributedRequests(reason)).toEqual([
      { api: "QueryService/metricsViewAggregation" },
    ]);
    expect(getAttributedRequests(thrown)).toEqual([
      { api: "QueryService/metricsViewAggregation" },
    ]);
  });

  it("lists every request that shared an abort reason, up to a limit", () => {
    const controller = new AbortController();
    const interceptor = createRequestAttributionInterceptor();
    for (const method of [
      "MetricsViewAggregation",
      "MetricsViewTimeSeries",
      "MetricsViewTotals",
      "MetricsViewRows",
      "MetricsViewSearch",
      "MetricsViewSchema",
    ]) {
      void interceptor(pending)(request(method, controller.signal));
    }
    const reason = new Error("signal is aborted without reason");

    controller.abort(reason);

    expect(getAttributedRequests(reason)?.map((r) => r.api)).toEqual([
      "QueryService/metricsViewAggregation",
      "QueryService/metricsViewTimeSeries",
      "QueryService/metricsViewTotals",
      "QueryService/metricsViewRows",
      "QueryService/metricsViewSearch",
    ]);
  });

  it("stops listening once the request has succeeded", async () => {
    const controller = new AbortController();
    const response = {} as UnaryResponse;

    await expect(
      createRequestAttributionInterceptor()(() => Promise.resolve(response))(
        request("MetricsViewAggregation", controller.signal),
      ),
    ).resolves.toBe(response);
    const reason = new Error("aborted after the call finished");
    controller.abort(reason);

    expect(getAttributedRequests(reason)).toBeUndefined();
  });

  it("tags the error that leaves the Connect transport when a request is aborted", async () => {
    const controller = new AbortController();
    const call = queryServiceClient(fetchUntilAborted).metricsViewAggregation(
      { instanceId: "inst-1", metricsView: "revenue" },
      { signal: controller.signal },
    );
    const reason = abortError("signal is aborted without reason");

    controller.abort(reason);
    const error = await call.catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ConnectError);
    expect(error).not.toBe(reason);
    expect((error as ConnectError).code).toBe(Code.Canceled);
    expect(getAttributedRequests(error)).toEqual([
      { api: "QueryService/metricsViewAggregation", resource: "revenue" },
    ]);
    expect(getAttributedRequests(reason)).toEqual([
      { api: "QueryService/metricsViewAggregation", resource: "revenue" },
    ]);
  });

  it("tags the error that leaves the Connect transport when a request fails", async () => {
    const failure = new TypeError("Failed to fetch");

    const error = await queryServiceClient(() => Promise.reject(failure))
      .metricsViewTotals({ instanceId: "inst-1", metricsViewName: "revenue" })
      .catch((e: unknown) => e);

    expect(error).toBeInstanceOf(ConnectError);
    expect((error as ConnectError).cause).toBe(failure);
    expect(getAttributedRequests(error)).toEqual([
      { api: "QueryService/metricsViewTotals", resource: "revenue" },
    ]);
  });

  it("knows nothing about errors that did not leave a runtime request", () => {
    expect(getAttributedRequests(new Error("boom"))).toBeUndefined();
    expect(getAttributedRequests("AbortError")).toBeUndefined();
    expect(getAttributedRequests(undefined)).toBeUndefined();
  });
});
