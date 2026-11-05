import { Code, ConnectError, type UnaryRequest } from "@connectrpc/connect";
import type { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import {
  type CommonUserFields,
  MetricsEventScreenName,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import { createRequestAttributionInterceptor } from "@rilldata/web-common/runtime-client/v2/request-attribution";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2/runtime-client";
import { CancelledError, type Query } from "@tanstack/query-core";
import { afterEach, beforeAll, describe, expect, it, vi } from "vitest";
import { ErrorEventHandler, getApiNameFromQueryKey } from "./ErrorEventHandler";
import { installResizeObserverTracker } from "./resize-observer-tracker";

const { PAGE_URL } = vi.hoisted(() => ({
  PAGE_URL: "https://ui.rilldata.com/acme/sales/explore/revenue",
}));

vi.mock("$app/stores", () => ({
  page: {
    subscribe: (run: (value: { url: URL }) => void) => {
      run({ url: new URL(PAGE_URL) });
      return () => {};
    },
  },
}));

const COMMON_USER_FIELDS = { locale: "en-US" } as CommonUserFields;

function createHandler() {
  const dispatch = vi.fn().mockResolvedValue(undefined);
  const handler = new ErrorEventHandler(
    { dispatch } as unknown as MetricsService,
    COMMON_USER_FIELDS,
    false,
    () => MetricsEventScreenName.Dashboard,
  );
  return { handler, dispatch };
}

function queryWithKey(queryKey: unknown[]): Query {
  return { queryKey } as unknown as Query;
}

function resizeEntries(...targets: Element[]): ResizeObserverEntry[] {
  return targets.map((target) => ({
    target,
  })) as unknown as ResizeObserverEntry[];
}

// In browsers the default abort reason is a DOMException that extends Error; jsdom's does not, so build it as browsers see it
function abortError(message: string): Error {
  const error = new Error(message);
  error.name = "AbortError";
  return error;
}

/** Starts a runtime RPC through the attribution interceptor and aborts it, returning the reason as it surfaces. */
function abortRuntimeRequest(
  serviceTypeName: string,
  methodName: string,
  message: Record<string, unknown> = {},
): Error {
  const controller = new AbortController();
  const request = {
    service: { typeName: serviceTypeName },
    method: { name: methodName },
    message,
    signal: controller.signal,
  } as unknown as UnaryRequest;
  void createRequestAttributionInterceptor()(() => new Promise(() => {}))(
    request,
  );
  const reason = abortError("signal is aborted without reason");
  controller.abort(reason);
  return reason;
}

describe("getApiNameFromQueryKey", () => {
  it("names ConnectRPC runtime queries by service and method", () => {
    expect(
      getApiNameFromQueryKey([
        "QueryService",
        "metricsViewAggregation",
        "inst-1",
        { metricsView: "revenue", measures: [{ name: "total" }] },
      ]),
    ).toBe("QueryService/metricsViewAggregation");
    expect(
      getApiNameFromQueryKey([
        "RuntimeService",
        "getResource",
        "inst-1",
        { name: { kind: "rill.runtime.v1.Explore", name: "revenue" } },
      ]),
    ).toBe("RuntimeService/getResource");
    expect(
      getApiNameFromQueryKey(["ConnectorService", "oLAPListTables", "inst-1"]),
    ).toBe("ConnectorService/oLAPListTables");
  });

  it("names infinite ConnectRPC queries like their plain query", () => {
    expect(
      getApiNameFromQueryKey([
        "RuntimeService",
        "listResources",
        "inst-1",
        {},
        "infinite",
      ]),
    ).toBe("RuntimeService/listResources");
  });

  it("keeps admin REST paths as they are", () => {
    expect(
      getApiNameFromQueryKey([
        "/v1/orgs/acme/projects/sales",
        { issues: true },
      ]),
    ).toBe("/v1/orgs/acme/projects/sales");
    expect(getApiNameFromQueryKey(["/v1/orgs"])).toBe("/v1/orgs");
    expect(getApiNameFromQueryKey(["infinite", "/v1/orgs/acme/members"])).toBe(
      "infinite",
    );
  });

  it("keeps custom query keys as they are", () => {
    expect(getApiNameFromQueryKey(["dashboardFilteredRowsCt", "revenue"])).toBe(
      "dashboardFilteredRowsCt",
    );
    expect(
      getApiNameFromQueryKey([
        "metrics_sql",
        "inst-1",
        "chart",
        0,
        "select 1",
        "{}",
      ]),
    ).toBe("metrics_sql");
  });

  it("returns an empty name for a key that does not start with a string", () => {
    expect(getApiNameFromQueryKey([{ kind: "custom" }])).toBe("");
    expect(getApiNameFromQueryKey([])).toBe("");
  });
});

describe("ErrorEventHandler.requestErrorEventHandler", () => {
  it("reports the RPC method, status and message of a failed runtime query", () => {
    const { handler, dispatch } = createHandler();

    handler.requestErrorEventHandler(
      new ConnectError("controller: resource not found", Code.NotFound),
      queryWithKey([
        "RuntimeService",
        "getResource",
        "inst-1",
        { name: { kind: "rill.runtime.v1.Explore", name: "revenue" } },
      ]),
    );

    expect(dispatch).toHaveBeenCalledWith("httpErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      "RuntimeService/getResource",
      "404",
      "controller: resource not found",
      PAGE_URL,
    ]);
  });

  it("reports admin REST errors under their path", () => {
    const { handler, dispatch } = createHandler();

    handler.requestErrorEventHandler(
      { response: { status: 403, data: { message: "not allowed" } } },
      queryWithKey(["/v1/orgs/acme/projects/sales", { issues: true }]),
    );

    expect(dispatch).toHaveBeenCalledWith("httpErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      "/v1/orgs/acme/projects/sales",
      "403",
      "not allowed",
      PAGE_URL,
    ]);
  });

  it("reports cancelled requests with the reason the client gave", () => {
    const { handler, dispatch } = createHandler();

    handler.requestErrorEventHandler(
      ConnectError.from(
        new DOMException(
          "Request cancelled: runtime client disposed",
          "AbortError",
        ),
      ),
      queryWithKey(["QueryService", "metricsViewAggregation", "inst-1", {}]),
    );

    expect(dispatch).toHaveBeenCalledWith("httpErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      "QueryService/metricsViewAggregation",
      expect.stringMatching(/^(408|500)$/),
      expect.stringContaining("Request cancelled: runtime client disposed"),
      PAGE_URL,
    ]);
  });

  it("reports timeouts and server failures that mention cancellation", () => {
    const { handler, dispatch } = createHandler();
    const hungReconcile =
      'controller event loop failed: reconciles for resources [kind:"rill.runtime.v1.Model" name:"orders"] have hung for more than 10m0s after cancelation';

    handler.requestErrorEventHandler(
      new ConnectError("context deadline exceeded", Code.DeadlineExceeded),
      queryWithKey(["QueryService", "metricsViewAggregation", "inst-1", {}]),
    );
    handler.requestErrorEventHandler(
      new ConnectError(hungReconcile, Code.Unknown),
      queryWithKey(["RuntimeService", "getResource", "inst-1", {}]),
    );

    expect(dispatch).toHaveBeenNthCalledWith(1, "httpErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      "QueryService/metricsViewAggregation",
      "408",
      "context deadline exceeded",
      PAGE_URL,
    ]);
    expect(dispatch).toHaveBeenNthCalledWith(2, "httpErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      "RuntimeService/getResource",
      "500",
      hungReconcile,
      PAGE_URL,
    ]);
  });
});

describe("ErrorEventHandler.addJavascriptErrorListeners", () => {
  let removeListeners: (() => void) | undefined;
  let lastObserverCallback: ResizeObserverCallback | undefined;

  beforeAll(() => {
    // jsdom has no ResizeObserver: a stub whose callback the test can invoke, wrapped by the tracker
    class StubResizeObserver {
      constructor(callback: ResizeObserverCallback) {
        lastObserverCallback = callback;
      }
      observe() {}
      unobserve() {}
      disconnect() {}
    }
    Object.defineProperty(window, "ResizeObserver", {
      configurable: true,
      writable: true,
      value: StubResizeObserver,
    });
    installResizeObserverTracker();
  });

  afterEach(() => {
    removeListeners?.();
    removeListeners = undefined;
    vi.unstubAllGlobals();
  });

  function dispatchErrorEvent(init: ErrorEventInit) {
    window.dispatchEvent(new ErrorEvent("error", init));
  }

  // jsdom has no PromiseRejectionEvent; the handler only reads `reason`.
  function dispatchRejection(reason: unknown) {
    window.dispatchEvent(
      Object.assign(new Event("unhandledrejection"), { reason }),
    );
  }

  it("reports uncaught errors", () => {
    const { handler, dispatch } = createHandler();
    removeListeners = handler.addJavascriptErrorListeners();
    const error = new TypeError(
      "Cannot read properties of undefined (reading 'type')",
    );

    dispatchErrorEvent({
      message: `Uncaught ${error.message}`,
      error,
      filename: "https://ui.rilldata.com/_app/immutable/chunks/app.js",
    });

    expect(dispatch).toHaveBeenCalledWith("javascriptErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      error.stack,
      `Uncaught ${error.message}`,
      PAGE_URL,
      "",
    ]);
  });

  it("reports unhandled rejections", () => {
    const { handler, dispatch } = createHandler();
    removeListeners = handler.addJavascriptErrorListeners();
    const error = new Error("Unrecognized data set: metrics-view");

    dispatchRejection(error);

    expect(dispatch).toHaveBeenCalledWith("javascriptErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      error.stack,
      error.message,
      PAGE_URL,
      "",
    ]);
  });

  it("names the runtime request behind an unhandled abort", () => {
    const { handler, dispatch } = createHandler();
    removeListeners = handler.addJavascriptErrorListeners();
    const reason = abortRuntimeRequest(
      "rill.runtime.v1.QueryService",
      "MetricsViewAggregation",
      { metricsView: "revenue" },
    );

    dispatchRejection(reason);

    expect(dispatch).toHaveBeenCalledWith("javascriptErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      expect.stringContaining(
        "request: QueryService/metricsViewAggregation (revenue)",
      ),
      "signal is aborted without reason [aborted request: QueryService/metricsViewAggregation]",
      PAGE_URL,
      "QueryService/metricsViewAggregation",
    ]);
  });

  it("names the runtime request behind an unhandled abort of a runtime client call", async () => {
    const { handler, dispatch } = createHandler();
    removeListeners = handler.addJavascriptErrorListeners();
    // Like a browser fetch: rejects with the signal's reason once it aborts
    const fetchMock = vi.fn<typeof fetch>(
      (_input: RequestInfo | URL, init?: RequestInit) =>
        new Promise<Response>((_, reject) => {
          const signal = init?.signal;
          if (signal?.aborted) reject(signal.reason);
          signal?.addEventListener("abort", () => reject(signal.reason));
        }),
    );
    vi.stubGlobal("fetch", fetchMock);
    const client = new RuntimeClient({
      host: "http://localhost:9009",
      instanceId: "inst-1",
    });
    const controller = new AbortController();
    const call = client.queryService.metricsViewAggregation(
      { instanceId: "inst-1", metricsView: "revenue" },
      { signal: controller.signal },
    );
    await vi.waitFor(() => expect(fetchMock).toHaveBeenCalled());

    controller.abort(abortError("signal is aborted without reason"));
    const rejection = await call.catch((e: unknown) => e);
    dispatchRejection(rejection);

    expect(rejection).toBeInstanceOf(ConnectError);
    expect(dispatch).toHaveBeenCalledWith("javascriptErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      expect.stringMatching(
        /\nrequest: QueryService\/metricsViewAggregation \(revenue\)$/,
      ),
      "[canceled] signal is aborted without reason [aborted request: QueryService/metricsViewAggregation]",
      PAGE_URL,
      "QueryService/metricsViewAggregation",
    ]);
  });

  it("names the runtime request behind an unhandled failure", () => {
    const { handler, dispatch } = createHandler();
    removeListeners = handler.addJavascriptErrorListeners();
    const error = new ConnectError("jwt is expired", Code.Unauthenticated);
    const request = {
      service: { typeName: "rill.runtime.v1.RuntimeService" },
      method: { name: "GetResource" },
      message: {},
      signal: new AbortController().signal,
    } as unknown as UnaryRequest;
    void createRequestAttributionInterceptor()(() => Promise.reject(error))(
      request,
    ).catch(() => {});

    return Promise.resolve().then(() => {
      dispatchRejection(error);

      expect(dispatch).toHaveBeenCalledWith("javascriptErrorEvent", [
        COMMON_USER_FIELDS,
        MetricsEventScreenName.Dashboard,
        expect.stringContaining("request: RuntimeService/getResource"),
        "[unauthenticated] jwt is expired [runtime request: RuntimeService/getResource]",
        PAGE_URL,
        "RuntimeService/getResource",
      ]);
    });
  });

  it("still reports aborts it cannot attribute, marked as such", () => {
    const { handler, dispatch } = createHandler();
    removeListeners = handler.addJavascriptErrorListeners();

    dispatchRejection(abortError("The user aborted a request."));
    dispatchRejection(new CancelledError());

    expect(dispatch).toHaveBeenNthCalledWith(1, "javascriptErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      expect.any(String),
      "The user aborted a request. [aborted request: no runtime request tagged]",
      PAGE_URL,
      "",
    ]);
    expect(dispatch).toHaveBeenNthCalledWith(2, "javascriptErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      expect.any(String),
      "CancelledError [aborted request: no runtime request tagged]",
      PAGE_URL,
      "",
    ]);
  });

  it("reports ResizeObserver loop notices with the callbacks that ran before them", () => {
    const { handler, dispatch } = createHandler();
    removeListeners = handler.addJavascriptErrorListeners();
    const callback = vi.fn();
    const observer = new ResizeObserver(callback);
    const table = document.createElement("table");
    table.id = "pivot-table";
    const canvas = document.createElement("div");
    canvas.id = "canvas";
    const row = document.createElement("div");
    row.className = "canvas-row flex h-full w-full";
    canvas.append(row);
    document.body.append(table, canvas);
    const entries = resizeEntries(table, row);

    lastObserverCallback?.(entries, observer);
    dispatchErrorEvent({
      message: "ResizeObserver loop completed with undelivered notifications.",
    });

    expect(callback).toHaveBeenCalledWith(entries, observer);
    expect(dispatch).toHaveBeenCalledWith("javascriptErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      expect.stringMatching(
        /^-?\d+ms: table#pivot-table, div\.canvas-row\.flex\.h-full in #canvas$/m,
      ),
      "ResizeObserver loop completed with undelivered notifications. [last callbacks: table#pivot-table, div.canvas-row.flex.h-full in #canvas]",
      PAGE_URL,
      "",
    ]);
    table.remove();
    canvas.remove();
  });

  it("reports cross-origin script errors as the browser gives them", () => {
    const { handler, dispatch } = createHandler();
    removeListeners = handler.addJavascriptErrorListeners();
    const script = document.createElement("script");
    script.src = "https://cdn.example.com/widget.js?token=query-secret";
    document.head.append(script);

    dispatchErrorEvent({ message: "Script error." });

    expect(dispatch).toHaveBeenCalledWith("javascriptErrorEvent", [
      COMMON_USER_FIELDS,
      MetricsEventScreenName.Dashboard,
      "",
      "Script error.",
      PAGE_URL,
      "",
    ]);
    script.remove();
  });

  it("does not report an error the query cache already reported again when its rejection goes unhandled", () => {
    const { handler, dispatch } = createHandler();
    removeListeners = handler.addJavascriptErrorListeners();
    const error = new ConnectError(
      "controller: resource not found",
      Code.NotFound,
    );

    // The order TanStack Query uses: the cache's onError, then the rejection of an uncaught fetchQuery
    handler.requestErrorEventHandler(
      error,
      queryWithKey(["RuntimeService", "getResource", "inst-1", {}]),
    );
    dispatchRejection(error);
    dispatchErrorEvent({ message: `Uncaught ${error.message}`, error });

    expect(dispatch).toHaveBeenCalledTimes(1);
    expect(dispatch).toHaveBeenCalledWith(
      "httpErrorEvent",
      expect.arrayContaining(["RuntimeService/getResource", "404"]),
    );
  });
});
