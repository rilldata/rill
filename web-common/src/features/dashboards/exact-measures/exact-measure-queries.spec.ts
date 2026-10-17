import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { exploreExactSurface } from "./explore-surfaces";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { initStateManagers } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import { createTimeSeriesDataStore } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { beforeEach, describe, expect, it, vi } from "vitest";

// The bid price measure gets an exact variant; impressions stays single-precision.
// `bid_price_per_impression` is calculated from both; the explore does not show it.
const BID_PRICE_PER_IMPRESSION_MEASURE = "bid_price_per_impression";
const AD_BIDS_METRICS_WITH_EXACT: V1MetricsViewSpec = {
  ...AD_BIDS_METRICS_INIT_WITH_TIME,
  measures: [
    ...(AD_BIDS_METRICS_INIT_WITH_TIME.measures?.map((measure) =>
      measure.name === AD_BIDS_BID_PRICE_MEASURE
        ? { ...measure, exactExpression: "avg(bid_price)" }
        : measure,
    ) ?? []),
    {
      name: BID_PRICE_PER_IMPRESSION_MEASURE,
      expression: "bid_price / impressions",
      referencedMeasures: [
        AD_BIDS_BID_PRICE_MEASURE,
        AD_BIDS_IMPRESSIONS_MEASURE,
      ],
    },
  ],
};

describe("exact measures in explore queries", () => {
  const dashboardFetchMocks = DashboardFetchMocks.useDashboardFetchMocks();

  let requestBodies: string[] = [];

  beforeEach(() => {
    requestBodies = [];
    const inner = globalThis.fetch;
    vi.stubGlobal("fetch", (url: any, opts: any) => {
      const body = opts?.body;
      if (body) {
        requestBodies.push(
          typeof body === "string"
            ? body
            : new TextDecoder().decode(body as ArrayBufferView),
        );
      }
      return inner(url, opts);
    });

    dashboardFetchMocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_WITH_EXACT,
      AD_BIDS_EXPLORE_INIT,
    );
    dashboardFetchMocks.mockTimeRangeSummary(AD_BIDS_NAME, {
      min: "2022-01-01",
      max: "2022-03-31",
    });
    dashboardFetchMocks.mockMetricsViewTimeRanges(
      AD_BIDS_NAME,
      "2022-01-01T00:00:00Z",
      "2022-03-31T00:00:00Z",
    );
    dashboardFetchMocks.mockMetricsViewAggregation(
      new RegExp(`"name":"${AD_BIDS_PUBLISHER_DIMENSION}"`),
      {
        schema: { fields: [{ name: AD_BIDS_PUBLISHER_DIMENSION }] },
        data: [{ [AD_BIDS_PUBLISHER_DIMENSION]: "Google" }],
      },
    );
  });

  it("flags only the selected measure and resets after a filter change", async () => {
    const { stateManagers, destroy } = initStateManagers();

    stateManagers.actions.exactMeasures.toggleExactMeasure(
      exploreExactSurface.measures,
      AD_BIDS_BID_PRICE_MEASURE,
    );

    const unsub = createTimeSeriesDataStore(stateManagers).subscribe(() => {});
    await asyncWait(500);

    const aggregationBodies = requestBodies.filter(
      (body) =>
        body.includes(`"name":"${AD_BIDS_BID_PRICE_MEASURE}"`) &&
        !body.includes("measureNames"),
    );
    expect(aggregationBodies.length).toBeGreaterThan(0);
    for (const body of aggregationBodies) {
      expect(body).toMatch(
        new RegExp(`"name":"${AD_BIDS_BID_PRICE_MEASURE}","exact":true`),
      );
      expect(body).not.toMatch(
        new RegExp(`"name":"${AD_BIDS_IMPRESSIONS_MEASURE}","exact":true`),
      );
    }

    const timeSeriesBodies = requestBodies.filter((body) =>
      body.includes("measureNames"),
    );
    expect(timeSeriesBodies.length).toBeGreaterThan(0);
    for (const body of timeSeriesBodies) {
      expect(body).toContain(
        `"exactMeasureNames":["${AD_BIDS_BID_PRICE_MEASURE}"]`,
      );
    }

    // A query context change (here: comparing by a dimension) resets the
    // selection: the next queries are approximate.
    requestBodies = [];
    metricsExplorerStore.setComparisonDimension(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_PUBLISHER_DIMENSION,
    );
    await asyncWait(500);

    expect(requestBodies.length).toBeGreaterThan(0);
    expect(
      requestBodies.filter((body) => body.includes('"exact":true')),
    ).toEqual([]);
    expect(
      requestBodies.filter((body) => body.includes("exactMeasureNames")),
    ).toEqual([]);

    unsub();
    destroy();
  });

  it("flags the exact measure a selected derived measure is calculated from in every request", async () => {
    const { stateManagers, destroy } = initStateManagers();

    // The derived measure is not shown on the explore, so no request names it.
    // Its input is exact nonetheless, because the runtime would compute the
    // input exactly wherever the derived measure is requested exactly.
    stateManagers.actions.exactMeasures.toggleExactMeasure(
      exploreExactSurface.measures,
      BID_PRICE_PER_IMPRESSION_MEASURE,
    );

    const unsub = createTimeSeriesDataStore(stateManagers).subscribe(() => {});
    await asyncWait(500);

    const aggregationBodies = requestBodies.filter(
      (body) =>
        body.includes(`"name":"${AD_BIDS_BID_PRICE_MEASURE}"`) &&
        !body.includes("measureNames"),
    );
    expect(aggregationBodies.length).toBeGreaterThan(0);
    for (const body of aggregationBodies) {
      expect(body).not.toContain(BID_PRICE_PER_IMPRESSION_MEASURE);
      expect(body).toMatch(
        new RegExp(`"name":"${AD_BIDS_BID_PRICE_MEASURE}","exact":true`),
      );
      expect(body).not.toMatch(
        new RegExp(`"name":"${AD_BIDS_IMPRESSIONS_MEASURE}","exact":true`),
      );
    }

    const timeSeriesBodies = requestBodies.filter((body) =>
      body.includes("measureNames"),
    );
    expect(timeSeriesBodies.length).toBeGreaterThan(0);
    for (const body of timeSeriesBodies) {
      expect(body).toContain(
        `"exactMeasureNames":["${AD_BIDS_BID_PRICE_MEASURE}"]`,
      );
    }

    unsub();
    destroy();
  });

  it("keeps the time series approximate for a selection made in a leaderboard", async () => {
    const { stateManagers, destroy } = initStateManagers();

    stateManagers.actions.exactMeasures.toggleExactMeasure(
      exploreExactSurface.leaderboard(AD_BIDS_PUBLISHER_DIMENSION),
      AD_BIDS_BID_PRICE_MEASURE,
    );

    const unsub = createTimeSeriesDataStore(stateManagers).subscribe(() => {});
    await asyncWait(500);

    expect(requestBodies.length).toBeGreaterThan(0);
    expect(
      requestBodies.filter((body) => body.includes('"exact":true')),
    ).toEqual([]);
    expect(
      requestBodies.filter((body) => body.includes("exactMeasureNames")),
    ).toEqual([]);

    unsub();
    destroy();
  });
});
