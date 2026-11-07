import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { initStateManagers } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers";
import {
  AD_BIDS_APPLY_PUB_DIMENSION_FILTER,
  applyMutationsToDashboard,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations";
import { createTimeDimensionDataStore } from "@rilldata/web-common/features/dashboards/time-dimension-details/time-dimension-data-store";
import { createTimeSeriesDataStore } from "@rilldata/web-common/features/dashboards/time-series/timeseries-data-store";
import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import { beforeEach, describe, expect, it, vi } from "vitest";

const EPHEMERAL_MEASURE = {
  name: "bid_price_per_impression",
  displayName: "Bid price per impression",
  expression: `${AD_BIDS_BID_PRICE_MEASURE} / ${AD_BIDS_IMPRESSIONS_MEASURE}`,
};

describe("ephemeral measures in explore queries", () => {
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
      AD_BIDS_METRICS_INIT_WITH_TIME,
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
        data: [
          { [AD_BIDS_PUBLISHER_DIMENSION]: "Google" },
          { [AD_BIDS_PUBLISHER_DIMENSION]: "Facebook" },
        ],
      },
    );
  });

  it("keeps the definition attached in the TDD dimension comparison", async () => {
    await runScenario({ expandEphemeralMeasureInTdd: true });
    expectEveryReferenceCarriesTheDefinition(requestBodies);
  });

  it("keeps the definition attached in the explore dimension comparison", async () => {
    await runScenario({ expandEphemeralMeasureInTdd: false });
    expectEveryReferenceCarriesTheDefinition(requestBodies);
  });

  async function runScenario({
    expandEphemeralMeasureInTdd,
  }: {
    expandEphemeralMeasureInTdd: boolean;
  }) {
    const { stateManagers, destroy } = initStateManagers();

    metricsExplorerStore.addEphemeralMeasure(
      AD_BIDS_EXPLORE_NAME,
      EPHEMERAL_MEASURE,
    );
    metricsExplorerStore.setComparisonDimension(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_PUBLISHER_DIMENSION,
    );
    if (expandEphemeralMeasureInTdd) {
      metricsExplorerStore.setExpandedMeasureName(
        AD_BIDS_EXPLORE_NAME,
        EPHEMERAL_MEASURE.name,
      );
    } else {
      // The explore chart reads its comparison values from the dimension
      // filter, so one has to be applied for those queries to run.
      await applyMutationsToDashboard(
        AD_BIDS_EXPLORE_NAME,
        [AD_BIDS_APPLY_PUB_DIMENSION_FILTER],
        stateManagers.expressionFilterManager,
      );
    }

    const unsubs = [
      createTimeSeriesDataStore(stateManagers).subscribe(() => {}),
      createTimeDimensionDataStore(stateManagers).subscribe(() => {}),
    ];
    await asyncWait(500);
    unsubs.forEach((u) => u());
    destroy();
  }
});

/**
 * A request is broken if it names the ephemeral measure but carries neither the
 * `expression` compute (aggregation requests) nor an `ephemeralMeasures` entry
 * (time series requests): the runtime then resolves the name against the metrics
 * view and fails with `measure "..." not found`.
 */
function expectEveryReferenceCarriesTheDefinition(bodies: string[]) {
  const referencing = bodies.filter((body) =>
    body.includes(EPHEMERAL_MEASURE.name),
  );
  expect(referencing.length).toBeGreaterThan(0);
  expect(
    referencing.filter((body) => !body.includes(EPHEMERAL_MEASURE.expression)),
  ).toEqual([]);
}
