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
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
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
  type V1ExploreSpec,
  type V1MetricsViewAggregationRequest,
  type V1MetricsViewSpec,
  V1Operation,
  V1TimeGrain,
  type V1TimeRangeSummary,
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
            showTotalsColumn: true,
            showTotalsRow: true,
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
            showTotalsColumn: true,
            showTotalsRow: true,
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
            showTotalsColumn: true,
            showTotalsRow: true,
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
            showTotalsColumn: true,
            showTotalsRow: true,
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

  // Reports and alerts embed the ephemeral definition in the request itself,
  // so opening one must restore it without any saved dashboard state.
  it("Recovers ephemeral measures from the request", async () => {
    const ephemeralMeasureName = "bid_price_per_impression";
    const aggregationRequest: V1MetricsViewAggregationRequest = {
      dimensions: [],
      measures: [
        {
          name: ephemeralMeasureName,
          expression: {
            expression: `${AD_BIDS_BID_PRICE_MEASURE} / ${AD_BIDS_IMPRESSIONS_MEASURE}`,
            displayName: "Bid price per impression",
          },
        },
      ],
      sort: [{ desc: true, name: ephemeralMeasureName }],
    };

    const exploreState = await getExploreState(aggregationRequest, {
      forceOpenPivot: false,
    });
    expect(exploreState.ephemeralMeasures).toEqual([
      {
        name: ephemeralMeasureName,
        displayName: "Bid price per impression",
        expression: `${AD_BIDS_BID_PRICE_MEASURE} / ${AD_BIDS_IMPRESSIONS_MEASURE}`,
      },
    ]);
    expect(exploreState.visibleMeasures).toEqual([ephemeralMeasureName]);
    // The ephemeral measure is not in the explore spec, so it must not make
    // the dashboard think every spec measure is visible.
    expect(exploreState.allMeasuresVisible).toBe(false);

    const pivotState = await getExploreState(aggregationRequest, {
      forceOpenPivot: true,
    });
    expect(pivotState.ephemeralMeasures).toEqual(
      exploreState.ephemeralMeasures,
    );
    expect(pivotState.pivot.columns).toEqual([
      {
        id: ephemeralMeasureName,
        title: "Bid price per impression",
        type: PivotChipType.Measure,
      },
    ]);
  });

  describe("metrics view without a time dimension", () => {
    let fetchSpy: MockInstance<typeof fetch>;

    beforeEach(() => {
      mocks.mockMetricsView(NO_TIME_SOURCE.metricsViewName, NO_TIME_METRICS);
      mocks.mockMetricsExplore(
        NO_TIME_SOURCE.exploreName,
        NO_TIME_METRICS,
        NO_TIME_EXPLORE,
      );
      fetchSpy = vi.spyOn(globalThis, "fetch");
    });

    afterEach(() => {
      fetchSpy.mockRestore();
    });

    // The runtime rejects a time range summary request for this metrics view.
    function expectNoTimeRangeSummaryRequest() {
      expect(
        fetchSpy.mock.calls.filter(([input]) =>
          (input instanceof Request ? input.url : input.toString()).endsWith(
            "/MetricsViewTimeRange",
          ),
        ),
      ).toEqual([]);
    }

    const where = createAndExpression([
      createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Yahoo"]),
    ]);
    const TestCases: {
      title: string;
      aggregationRequest: V1MetricsViewAggregationRequest;
      expectedNonPivotState: Partial<ExploreState>;
      expectedPivotState: Partial<ExploreState>;
    }[] = [
      {
        title: "With a dimension, measure and filter",
        aggregationRequest: {
          dimensions: [{ name: AD_BIDS_DOMAIN_DIMENSION }],
          measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
          sort: [{ desc: true, name: AD_BIDS_BID_PRICE_MEASURE }],
          where,
        },
        expectedNonPivotState: {
          activePage: DashboardState_ActivePage.DIMENSION_TABLE,
          allMeasuresVisible: false,
          visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
          selectedDimensionName: AD_BIDS_DOMAIN_DIMENSION,
          leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
          whereFilter: where,
        },
        expectedPivotState: {
          activePage: DashboardState_ActivePage.PIVOT,
          whereFilter: where,
          pivot: {
            rows: [],
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
            showTotalsColumn: true,
            showTotalsRow: true,
            tableMode: "flat",
          },
        },
      },

      {
        title: "With only a single measure",
        aggregationRequest: {
          dimensions: [],
          measures: [{ name: AD_BIDS_BID_PRICE_MEASURE }],
          sort: [{ desc: true, name: AD_BIDS_BID_PRICE_MEASURE }],
        },
        // Time dimension details are not opened without a time dimension
        expectedNonPivotState: {
          allMeasuresVisible: false,
          visibleMeasures: [AD_BIDS_BID_PRICE_MEASURE],
          leaderboardSortByMeasureName: AD_BIDS_BID_PRICE_MEASURE,
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
            showTotalsColumn: true,
            showTotalsRow: true,
            tableMode: "flat",
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
          source: NO_TIME_SOURCE,
        });
        expectNoTimeRangeSummaryRequest();
      });

      it(`${title} : pivot state`, async () => {
        await runTest({
          aggregationRequest,
          expectedAdditionalExploreState: expectedPivotState,
          ignoreFilters: false,
          forceOpenPivot: true,
          source: NO_TIME_SOURCE,
        });
        expectNoTimeRangeSummaryRequest();
      });
    }
  });

  // TODO: add more extensive tests for other parts
});

async function getExploreState(
  aggregationRequest: V1MetricsViewAggregationRequest,
  { forceOpenPivot }: { forceOpenPivot: boolean },
) {
  const mockClient = new RuntimeClient({
    host: "http://localhost:9009",
    instanceId: "default",
  });
  const mapQueryStore = mapQueryToDashboard(
    mockClient,
    {
      exploreName: AD_BIDS_EXPLORE_NAME,
      queryName: "MetricsViewAggregation",
      queryArgsJson: JSON.stringify({
        metricsView: AD_BIDS_METRICS_NAME,
        ...aggregationRequest,
      }),
      executionTime: AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary!.max!,
    },
    { ignoreFilters: false, forceOpenPivot },
  );

  let mapQueryResp: MapQueryResponse | undefined;
  const unsub = mapQueryStore.subscribe((r) => (mapQueryResp = r));
  await waitUntil(() => !!mapQueryResp?.data, 1000, 50);
  unsub();

  if (!mapQueryResp?.data) {
    throw new Error("mapQueryStore did not return a response");
  }
  return mapQueryResp.data.exploreState;
}

type TestSource = {
  metricsViewName: string;
  exploreName: string;
  metricsView: V1MetricsViewSpec;
  explore: V1ExploreSpec;
  timeRangeSummary: V1TimeRangeSummary | undefined;
};

const AD_BIDS_SOURCE: TestSource = {
  metricsViewName: AD_BIDS_METRICS_NAME,
  exploreName: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  explore: AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  timeRangeSummary: AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
};

const NO_TIME_METRICS: V1MetricsViewSpec = {
  displayName: AD_BIDS_METRICS_3_MEASURES_DIMENSIONS.displayName,
  table: AD_BIDS_METRICS_3_MEASURES_DIMENSIONS.table,
  measures: AD_BIDS_METRICS_3_MEASURES_DIMENSIONS.measures,
  dimensions: AD_BIDS_METRICS_3_MEASURES_DIMENSIONS.dimensions,
};
const NO_TIME_EXPLORE: V1ExploreSpec = {
  ...AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  metricsView: "AdBids_no_time_metrics",
};
const NO_TIME_SOURCE: TestSource = {
  metricsViewName: "AdBids_no_time_metrics",
  exploreName: "AdBids_no_time_explore",
  metricsView: NO_TIME_METRICS,
  explore: NO_TIME_EXPLORE,
  timeRangeSummary: undefined,
};

async function runTest({
  aggregationRequest,
  expectedAdditionalExploreState,
  ignoreFilters,
  forceOpenPivot,
  source = AD_BIDS_SOURCE,
}: {
  aggregationRequest: V1MetricsViewAggregationRequest;
  expectedAdditionalExploreState: Partial<ExploreState>;
  ignoreFilters: boolean;
  forceOpenPivot: boolean;
  source?: TestSource;
}) {
  const mockClient = new RuntimeClient({
    host: "http://localhost:9009",
    instanceId: "default",
  });
  const mapQueryStore = mapQueryToDashboard(
    mockClient,
    {
      exploreName: source.exploreName,
      queryName: "MetricsViewAggregation",
      queryArgsJson: JSON.stringify({
        metricsView: source.metricsViewName,
        ...aggregationRequest,
      }),
      executionTime: AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary!.max!,
    },
    {
      ignoreFilters,
      forceOpenPivot,
    },
  );

  // The store starts out undefined when its queries are already cached.
  const responses: (MapQueryResponse | undefined)[] = [];
  const unsub = mapQueryStore.subscribe((r) => responses.push(r));
  await waitUntil(() => !!responses.at(-1)?.data, 1000, 50);
  unsub();

  const mapQueryResp = responses.at(-1);
  if (!mapQueryResp) {
    throw new Error("mapQueryStore did not return a response");
  }

  // No response along the way, not just the last one, should carry an error.
  expect(responses.map((r) => r?.error).filter(Boolean)).toEqual([]);

  const rillDefaultExploreState = getRillDefaultExploreState(
    source.metricsView,
    source.explore,
    source.timeRangeSummary,
  );
  const exploreStateFromYAMLConfig = getExploreStateFromYAMLConfig(
    source.explore,
    source.timeRangeSummary,
    source.metricsView.smallestTimeGrain,
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
