import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks.ts";
import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
  createSubQueryExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { getExploreStateFromYAMLConfig } from "@rilldata/web-common/features/dashboards/stores/get-explore-state-from-yaml-config.ts";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state.ts";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
  AD_BIDS_TIMESTAMP_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types.ts";
import {
  type MapQueryResponse,
  mapQueryToDashboard,
} from "@rilldata/web-common/features/explore-mappers/map-to-explore.ts";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb.ts";
import {
  type V1MetricsViewAggregationRequest,
  V1Operation,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import { beforeEach, describe, expect, it } from "vitest";

describe("getDashboardFromAggregationRequest", () => {
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();

  beforeEach(() => {
    mocks.mockMetricsView(
      AD_BIDS_METRICS_NAME,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
    );
    mocks.mockMetricsExplore(
      AD_BIDS_EXPLORE_NAME,
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
      AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
    );
    mocks.mockTimeRangeSummary(
      AD_BIDS_METRICS_NAME,
      AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary!,
    );
    mocks.mockMetricsViewTimeRanges(
      AD_BIDS_METRICS_NAME,
      "2023-01-01T00:00:00Z",
      "2023-01-01T06:00:00Z",
    );
  });

  describe("active page and settings", () => {
    const TestCases: {
      title: string;
      aggregationRequest: V1MetricsViewAggregationRequest;
      expectedNonPivotState: Partial<ExploreState>;
      expectedPivotState: Partial<ExploreState>;
    }[] = [
      {
        title: "With only a single measure",
        aggregationRequest: {
          dimensions: [],
          measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
          sort: [{ desc: true, name: AD_BIDS_BID_PRICE_MEASURE }],
        },
        expectedNonPivotState: {
          activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
          allMeasuresVisible: false,
          visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
          leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
          tdd: {
            expandedMeasureName: AD_BIDS_BID_PRICE_MEASURE,
            chartType: TDDChart.DEFAULT,
            pinIndex: -1,
          },
        },
        expectedPivotState: {
          activePage: DashboardState_ActivePage.PIVOT,
          pivot: {
            rows: [],
            columns: [
              {
                id: AD_BIDS_BID_PRICE_MEASURE,
                title: AD_BIDS_BID_PRICE_MEASURE,
                type: PivotChipType.Measure,
              },
            ],
            sorting: [
              {
                desc: true,
                id: AD_BIDS_BID_PRICE_MEASURE,
              },
            ],
            expanded: {},
            columnPage: 1,
            rowPage: 1,
            enableComparison: true,
            activeCell: null,
            tableMode: "flat",
          },
        },
      },

      {
        title: "With only a single dimension",
        aggregationRequest: {
          dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
          measures: [],
          sort: [{ desc: true, name: AD_BIDS_DOMAIN_DIMENSION }],
        },
        expectedNonPivotState: {
          activePage: DashboardState_ActivePage.DIMENSION_TABLE,
          selectedDimensionName: AD_BIDS_DOMAIN_DIMENSION,
        },
        expectedPivotState: {
          activePage: DashboardState_ActivePage.PIVOT,
          pivot: {
            rows: [],
            columns: [
              {
                id: AD_BIDS_DOMAIN_DIMENSION,
                title: AD_BIDS_DOMAIN_DIMENSION,
                type: PivotChipType.Dimension,
              },
            ],
            sorting: [
              {
                desc: true,
                id: AD_BIDS_DOMAIN_DIMENSION,
              },
            ],
            expanded: {},
            columnPage: 1,
            rowPage: 1,
            enableComparison: true,
            activeCell: null,
            tableMode: "flat",
          },
        },
      },

      {
        title: "With simple and time dimension, single measure",
        aggregationRequest: {
          dimensions: [
            { name: AD_BIDS_DOMAIN_DIMENSION },
            {
              name: AD_BIDS_TIMESTAMP_DIMENSION,
              timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
            },
          ],
          measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
          sort: [{ desc: true, name: AD_BIDS_BID_PRICE_MEASURE }],
        },
        // Time dimension is ignored
        expectedNonPivotState: {
          activePage: DashboardState_ActivePage.DIMENSION_TABLE,
          allMeasuresVisible: false,
          visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
          selectedDimensionName: AD_BIDS_DOMAIN_DIMENSION,
          leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        },
        expectedPivotState: {
          activePage: DashboardState_ActivePage.PIVOT,
          pivot: {
            rows: [],
            columns: [
              {
                id: AD_BIDS_DOMAIN_DIMENSION,
                title: AD_BIDS_DOMAIN_DIMENSION,
                type: PivotChipType.Dimension,
              },
              {
                id: V1TimeGrain.TIME_GRAIN_WEEK,
                title: "week",
                type: PivotChipType.Time,
              },
              {
                id: AD_BIDS_BID_PRICE_MEASURE,
                title: AD_BIDS_BID_PRICE_MEASURE,
                type: PivotChipType.Measure,
              },
            ],
            sorting: [
              {
                desc: true,
                id: AD_BIDS_BID_PRICE_MEASURE,
              },
            ],
            expanded: {},
            columnPage: 1,
            rowPage: 1,
            enableComparison: true,
            activeCell: null,
            tableMode: "flat",
          },
        },
      },

      {
        title: "With simple and time dimension, single measure and pivot",
        aggregationRequest: {
          dimensions: [
            { name: AD_BIDS_DOMAIN_DIMENSION },
            {
              name: AD_BIDS_TIMESTAMP_DIMENSION,
              timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
            },
          ],
          measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
          sort: [{ desc: true, name: AD_BIDS_BID_PRICE_MEASURE }],
          pivotOn: [AD_BIDS_DOMAIN_DIMENSION],
        },
        // Pivot is ignored
        expectedNonPivotState: {
          activePage: DashboardState_ActivePage.DIMENSION_TABLE,
          allMeasuresVisible: false,
          visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
          selectedDimensionName: AD_BIDS_DOMAIN_DIMENSION,
          leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        },
        expectedPivotState: {
          activePage: DashboardState_ActivePage.PIVOT,
          pivot: {
            rows: [
              {
                id: V1TimeGrain.TIME_GRAIN_WEEK,
                title: "week",
                type: PivotChipType.Time,
              },
            ],
            columns: [
              {
                id: AD_BIDS_DOMAIN_DIMENSION,
                title: AD_BIDS_DOMAIN_DIMENSION,
                type: PivotChipType.Dimension,
              },
              {
                id: AD_BIDS_BID_PRICE_MEASURE,
                title: AD_BIDS_BID_PRICE_MEASURE,
                type: PivotChipType.Measure,
              },
            ],
            sorting: [
              {
                desc: true,
                id: AD_BIDS_BID_PRICE_MEASURE,
              },
            ],
            expanded: {},
            columnPage: 1,
            rowPage: 1,
            enableComparison: true,
            activeCell: null,
            tableMode: "nest",
          },
        },
      },
    ];

    for (const {
      title,
      aggregationRequest,
      expectedNonPivotState,
      expectedPivotState,
    } of TestCases) {
      it(`${title} : non-pivot state`, async () => {
        await runTest({
          aggregationRequest,
          expectedAdditionalExploreState: expectedNonPivotState,
          ignoreFilters: false,
          forceOpenPivot: false,
        });
      });

      it(`${title} : pivot state`, async () => {
        await runTest({
          aggregationRequest,
          expectedAdditionalExploreState: expectedPivotState,
          ignoreFilters: false,
          forceOpenPivot: true,
        });
      });
    }
  });

  it("Ignore filters", async () => {
    await runTest({
      aggregationRequest: {
        dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
        measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
        sort: [{ desc: true, name: AD_BIDS_BID_PRICE_MEASURE }],
        where: createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo", "Google"]),
          createSubQueryExpression(
            AD_BIDS_DOMAIN_DIMENSION,
            [AD_BIDS_BID_PRICE_MEASURE],
            createBinaryExpression(
              "AD_BIDS_BID_PRICE_MEASURE",
              V1Operation.OPERATION_GT,
              1000,
            ),
          ),
        ]),
      },
      expectedAdditionalExploreState: {
        activePage: DashboardState_ActivePage.DIMENSION_TABLE,
        allMeasuresVisible: false,
        visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
        selectedDimensionName: AD_BIDS_DOMAIN_DIMENSION,
        leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
        // No filters added
      },
      ignoreFilters: true,
      forceOpenPivot: false,
    });
  });

  // TODO: add more extensive tests for other parts
});

async function runTest({
  aggregationRequest,
  expectedAdditionalExploreState,
  ignoreFilters,
  forceOpenPivot,
}: {
  aggregationRequest: V1MetricsViewAggregationRequest;
  expectedAdditionalExploreState: Partial<ExploreState>;
  ignoreFilters: boolean;
  forceOpenPivot: boolean;
}) {
  const mapQueryStore = mapQueryToDashboard(
    {
      exploreName: AD_BIDS_EXPLORE_NAME,
      queryName: "MetricsViewAggregation",
      queryArgsJson: JSON.stringify({
        metricsView: AD_BIDS_METRICS_NAME,
        ...aggregationRequest,
      }),
      executionTime: AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary!.max!,
    },
    {
      ignoreFilters,
      forceOpenPivot,
    },
  );

  let mapQueryResp: MapQueryResponse | undefined;
  const unsub = mapQueryStore.subscribe((r) => (mapQueryResp = r));
  await waitUntil(() => !!mapQueryResp?.data, 1000, 50);
  unsub();

  if (!mapQueryResp) {
    throw new Error("mapQueryStore did not return a response");
  }

  expect(mapQueryResp.error).toBeNull();

  const rillDefaultExploreState = getRillDefaultExploreState(
    AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
    AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
    AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
  );
  const exploreStateFromYAMLConfig = getExploreStateFromYAMLConfig(
    AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
    AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
  );
  const expectedExploreState = {
    ...rillDefaultExploreState,
    ...exploreStateFromYAMLConfig,
    ...expectedAdditionalExploreState,
  };
  delete expectedExploreState.selectedTimeRange;
  if (mapQueryResp.data?.exploreState) {
    delete mapQueryResp.data.exploreState.selectedTimeRange;
  }

  expect(mapQueryResp.data?.exploreState).toEqual(expectedExploreState);
}
