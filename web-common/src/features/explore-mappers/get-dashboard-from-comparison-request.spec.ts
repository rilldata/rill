import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks.ts";
import { getSortType } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils.ts";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types.ts";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { getExploreStateFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config.ts";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state.ts";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_METRICS_INIT,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import {
  type MapQueryResponse,
  mapQueryToDashboard,
} from "@rilldata/web-common/features/explore-mappers/map-to-explore.ts";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import {
  V1MetricsViewComparisonMeasureType,
  type V1MetricsViewComparisonRequest,
} from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  type MockInstance,
  vi,
} from "vitest";

describe("getDashboardFromComparisonRequest", () => {
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();

  describe("metrics view without a time dimension", () => {
    let fetchSpy: MockInstance<typeof fetch>;

    beforeEach(() => {
      // AD_BIDS_METRICS_INIT has no time dimension.
      mocks.mockMetricsView(AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_INIT);
      mocks.mockMetricsExplore(
        AD_BIDS_EXPLORE_NAME,
        AD_BIDS_METRICS_INIT,
        AD_BIDS_EXPLORE_INIT,
      );
      fetchSpy = vi.spyOn(globalThis, "fetch");
    });

    afterEach(() => {
      fetchSpy.mockRestore();
    });

    it("With a dimension, measure and filter", async () => {
      const where = createAndExpression([
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
      ]);
      const comparisonRequest: V1MetricsViewComparisonRequest = {
        metricsViewName: AD_BIDS_METRICS_NAME,
        dimension: { name: AD_BIDS_DOMAIN_DIMENSION },
        measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
        sort: [
          {
            name: AD_BIDS_BID_PRICE_MEASURE,
            sortType:
              V1MetricsViewComparisonMeasureType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
            desc: true,
          },
        ],
        where,
      };

      const mapQueryStore = mapQueryToDashboard(
        new RuntimeClient({
          host: "http://localhost:9009",
          instanceId: "default",
        }),
        {
          exploreName: AD_BIDS_EXPLORE_NAME,
          queryName: "MetricsViewComparison",
          queryArgsJson: JSON.stringify(comparisonRequest),
        },
        {},
      );

      const responses: (MapQueryResponse | undefined)[] = [];
      const unsub = mapQueryStore.subscribe((r) => responses.push(r));
      await waitUntil(() => !!responses.at(-1)?.data, 1000, 50);
      unsub();

      expect(responses.map((r) => r?.error).filter(Boolean)).toEqual([]);
      // The runtime rejects a time range summary request for this metrics view.
      expect(
        fetchSpy.mock.calls.filter(([input]) =>
          (input instanceof Request ? input.url : input.toString()).endsWith(
            "/MetricsViewTimeRange",
          ),
        ),
      ).toEqual([]);
      expect(responses.at(-1)?.data?.exploreState).toEqual({
        ...getRillDefaultExploreState(
          AD_BIDS_METRICS_INIT,
          AD_BIDS_EXPLORE_INIT,
          undefined,
        ),
        ...getExploreStateFromYAMLConfig(
          AD_BIDS_EXPLORE_INIT,
          undefined,
          AD_BIDS_METRICS_INIT.smallestTimeGrain,
        ),
        whereFilter: where,
        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        sortDirection: SortDirection.DESCENDING,
        dashboardSortType: getSortType(
          V1MetricsViewComparisonMeasureType.METRICS_VIEW_COMPARISON_MEASURE_TYPE_BASE_VALUE,
        ),
        selectedDimensionName: AD_BIDS_DOMAIN_DIMENSION,
        activePage: DashboardState_ActivePage.DIMENSION_TABLE,
      });
    });
  });
});
