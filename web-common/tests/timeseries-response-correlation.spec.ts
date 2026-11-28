import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { EventEmitter } from "node:events";
import type { Page, Response } from "@playwright/test";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { interceptTimeseriesResponse } from "../../web-local/tests/utils/dataSpecifcHelpers";

// Only the response-listener helper is exercised; unrelated browser helpers stay unloaded.
vi.mock("@playwright/test", () => ({ expect: vi.fn() }));
vi.mock("../../web-local/tests/utils/metricsViewHelpers", () => ({
  assertLeaderboards: vi.fn(),
}));
vi.mock("../../web-local/tests/utils/commonHelpers", () => ({}));
vi.mock("../../web-local/tests/utils/modelHelpers", () => ({}));
vi.mock("../../web-local/tests/utils/sourceHelpers", () => ({}));

const expected = {
  metricsViewName: "AdBids_metrics",
  measureName: "total_records",
  timeGranularity: V1TimeGrain.TIME_GRAIN_DAY,
  timeZone: "America/Los_Angeles",
};
const request = {
  metricsViewName: expected.metricsViewName,
  measureNames: [expected.measureName],
  timeGranularity: expected.timeGranularity,
  timeZone: expected.timeZone,
};
const daily = {
  data: Array.from({ length: 7 }, () => ({ ts: "2022-03-24", records: {} })),
};
const weekly = {
  data: Array.from({ length: 14 }, () => ({ ts: "2022-01-01", records: {} })),
};

function response(
  body = daily,
  fields: Record<string, unknown> = {},
  options: {
    ok?: boolean;
    method?: string;
    url?: string;
    badRequest?: boolean;
    badBody?: boolean;
  } = {},
): Response {
  return {
    url: () =>
      options.url ??
      "http://localhost/rill.runtime.v1.QueryService/MetricsViewTimeSeries",
    ok: () => options.ok ?? true,
    request: () => ({
      method: () => options.method ?? "POST",
      postDataJSON: () => {
        if (options.badRequest) throw new Error("invalid request");
        return { ...request, ...fields };
      },
    }),
    json: async () => {
      if (options.badBody) throw new Error("body unavailable");
      return body;
    },
  } as Response;
}

describe("timeseries response correlation", () => {
  let events: EventEmitter;
  let page: Page;
  beforeEach(() => {
    vi.useFakeTimers();
    events = new EventEmitter();
    page = events as unknown as Page;
  });
  afterEach(() => vi.useRealTimers());

  it("ignores a stale weekly response before the requested daily response", async () => {
    const pending = interceptTimeseriesResponse(page, expected);
    const settled = vi.fn();
    void pending.then(settled);
    events.emit(
      "response",
      response(weekly, { timeGranularity: V1TimeGrain.TIME_GRAIN_WEEK }),
    );
    await vi.advanceTimersByTimeAsync(0);
    expect(settled).not.toHaveBeenCalled();
    events.emit("response", response(daily));
    expect(await pending).toEqual(daily);
    expect(events.listenerCount("response")).toBe(0);
    expect(vi.getTimerCount()).toBe(0);
  });

  it.each([
    { measureNames: ["bid_price_sum"] },
    { measureNames: ["total_records", "bid_price_sum"] },
    { metricsViewName: "another_view" },
    { timeZone: "UTC" },
  ])("ignores another request %j", async (fields) => {
    const pending = interceptTimeseriesResponse(page, expected);
    events.emit("response", response(weekly, fields));
    events.emit("response", response(daily));
    expect(await pending).toEqual(daily);
  });

  it.each([
    { ok: false },
    { method: "GET" },
    { url: "http://localhost/another-rpc" },
    { badRequest: true },
    { badBody: true },
  ])("keeps waiting after an unusable response %j", async (options) => {
    const pending = interceptTimeseriesResponse(page, expected);
    events.emit("response", response(weekly, {}, options));
    await vi.advanceTimersByTimeAsync(0);
    expect(events.listenerCount("response")).toBe(1);
    events.emit("response", response(daily));
    expect(await pending).toEqual(daily);
  });

  it("does not filter response length, so product data assertions still see wrong data", async () => {
    const pending = interceptTimeseriesResponse(page, expected);
    events.emit("response", response(weekly));
    expect(await pending).toEqual(weekly);
  });

  it("times out and detaches if no matching response arrives", async () => {
    const pending = interceptTimeseriesResponse(page, expected);
    events.emit(
      "response",
      response(weekly, { timeGranularity: V1TimeGrain.TIME_GRAIN_WEEK }),
    );
    await Promise.all([
      expect(pending).rejects.toThrow(
        "Timeout waiting for timeseries response",
      ),
      vi.advanceTimersByTimeAsync(30_000),
    ]);
    expect(events.listenerCount("response")).toBe(0);
    expect(vi.getTimerCount()).toBe(0);
  });
});
