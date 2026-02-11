import {
  aggregationRequestWithRowsAndColumns,
  buildAggregationRequest,
} from "@rilldata/web-common/features/dashboards/aggregation-request-utils.ts";
import { getDimensionTableAggregationRequestForTime } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-export.ts";
import { LeaderboardContextColumn } from "@rilldata/web-common/features/dashboards/leaderboard-context-column.ts";
import { splitPivotChips } from "@rilldata/web-common/features/dashboards/pivot/pivot-utils.ts";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores.ts";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT as AD_BIDS_EXPLORE,
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_PUBLISHER_DIMENSION,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import { getInitExploreStateForTest } from "@rilldata/web-common/features/dashboards/stores/test-data/helpers.ts";
import {
  AD_BIDS_FLAT_PIVOT_TABLE,
  AD_BIDS_OPEN_IMP_TDD,
  AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS,
  AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
  AD_BIDS_SET_DOMAIN_COMPARE_DIMENSION,
  AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
  AD_BIDS_SET_TIME_PIVOT_FILTER,
  AD_BIDS_SORT_ASC_BY_BID_PRICE,
  AD_BIDS_SORT_BY_PERCENT_CHANGE_IMPRESSIONS,
  applyMutationsToDashboard,
  type TestDashboardMutation,
} from "@rilldata/web-common/features/dashboards/stores/test-data/store-mutations.ts";
import {
  getTimeControlState,
  type TimeControlState,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { extractRowsAndColumns } from "@rilldata/web-common/features/scheduled-reports/utils.ts";
import type { V1MetricsViewAggregationRequest } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { describe, expect, it } from "vitest";
import { getPivotAggregationRequest } from "web-common/src/features/dashboards/pivot/pivot-export.ts";
import { getTDDAggregationRequest } from "web-common/src/features/dashboards/time-dimension-details/tdd-export.ts";

describe("Report rows and columns", () => {
  describe("From dimension table", () => {
    const testCases: TestCase[] = [
      {
        title: "Changing dimension and measure",
        mutations: [
          AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
          AD_BIDS_SORT_ASC_BY_BID_PRICE,
        ],
        expectedRows: [],
        expectedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
        ],

        updatedRows: [],
        updatedColumns: [AD_BIDS_DOMAIN_DIMENSION, AD_BIDS_IMPRESSIONS_MEASURE],

        expectedRequest: {
          dimensions: [{ name: "domain" }],
          measures: [{ name: "impressions" }],
          sort: [{ desc: true, name: "impressions" }],
        },
      },

      {
        title: "Adding row dimension",

        mutations: [
          AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
          AD_BIDS_SORT_ASC_BY_BID_PRICE,
        ],
        expectedRows: [],
        expectedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
        ],

        updatedRows: [AD_BIDS_DOMAIN_DIMENSION],
        updatedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
        ],

        expectedRequest: {
          dimensions: [{ name: "domain" }, { name: "publisher" }],
          measures: [{ name: "impressions" }, { name: "bid_price" }],
          sort: [{ desc: false, name: "domain" }],
          pivotOn: ["publisher"],
        },
      },

      {
        title: "Adding column dimension",

        mutations: [
          AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
          AD_BIDS_SORT_ASC_BY_BID_PRICE,
        ],
        expectedRows: [],
        expectedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
        ],

        updatedRows: [],
        updatedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
          AD_BIDS_DOMAIN_DIMENSION,
        ],

        expectedRequest: {
          dimensions: [{ name: "publisher" }, { name: "domain" }],
          measures: [{ name: "impressions" }, { name: "bid_price" }],
          sort: [{ desc: false, name: "bid_price" }],
        },
      },

      {
        title: "Adding row time dimension",

        mutations: [
          AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
          AD_BIDS_SORT_ASC_BY_BID_PRICE,
        ],
        expectedRows: [],
        expectedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
        ],

        updatedRows: ["timestamp_rill_TIME_GRAIN_HOUR"],
        updatedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
        ],

        expectedRequest: {
          dimensions: [
            {
              alias: "Time hour",
              name: "timestamp",
              timeGrain: "TIME_GRAIN_HOUR",
              timeZone: "UTC",
            },
            { name: "publisher" },
          ],
          measures: [{ name: "impressions" }, { name: "bid_price" }],
          sort: [{ desc: false, name: "Time hour" }],
          pivotOn: ["publisher"],
        },
      },

      {
        title: "Sorting by comparison measure",

        mutations: [
          AD_BIDS_SET_PREVIOUS_PERIOD_COMPARE_TIME_RANGE_FILTER,
          AD_BIDS_OPEN_PUB_DIMENSION_TABLE,
          AD_BIDS_SORT_BY_PERCENT_CHANGE_IMPRESSIONS,
        ],
        expectedRows: [],
        expectedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          AD_BIDS_IMPRESSIONS_MEASURE,
          AD_BIDS_BID_PRICE_MEASURE,
        ],

        updatedRows: [],
        updatedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        expectedRequest: {
          dimensions: [{ name: "publisher" }],
          measures: [
            { name: "impressions" },
            {
              comparisonValue: { measure: "impressions" },
              name: "impressions_prev",
            },
            {
              comparisonDelta: { measure: "impressions" },
              name: "impressions_delta",
            },
            {
              comparisonRatio: { measure: "impressions" },
              name: "impressions_delta_perc",
            },
          ],
          sort: [{ desc: true, name: "impressions_delta_perc" }],
        },
      },
    ];

    testCases.forEach((testCase) => {
      it(testCase.title, async () => {
        await runTest(testCase, (exploreState, timeControlState) =>
          getDimensionTableAggregationRequestForTime({
            metricsViewName: AD_BIDS_METRICS_NAME,
            exploreState,
            timeRange: {
              start: timeControlState.timeStart,
              end: timeControlState.timeEnd,
            },
            comparisonTimeRange: timeControlState.selectedComparisonTimeRange
              ? {
                  start: timeControlState.comparisonTimeStart,
                  end: timeControlState.comparisonTimeEnd,
                }
              : undefined,
            dimensionSearchText: "",
          }),
        );
      });
    });
  });

  describe("From TDD", () => {
    const testCases: TestCase[] = [
      {
        title: "Change grain",

        mutations: [AD_BIDS_OPEN_IMP_TDD, AD_BIDS_SET_DOMAIN_COMPARE_DIMENSION],
        expectedRows: [AD_BIDS_DOMAIN_DIMENSION],
        expectedColumns: [
          "timestamp_rill_TIME_GRAIN_HOUR",
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        updatedRows: [AD_BIDS_DOMAIN_DIMENSION],
        updatedColumns: [
          "timestamp_rill_TIME_GRAIN_DAY",
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        expectedRequest: {
          dimensions: [
            { name: "domain" },
            {
              name: "timestamp",
              timeGrain: "TIME_GRAIN_DAY",
              timeZone: "UTC",
              alias: "Time day",
            },
          ],
          measures: [{ name: "impressions" }],
          pivotOn: ["Time day"],
          sort: [{ name: "domain", desc: true }],
        },
      },

      {
        title: "Change dimension",

        mutations: [AD_BIDS_OPEN_IMP_TDD, AD_BIDS_SET_DOMAIN_COMPARE_DIMENSION],
        expectedRows: [AD_BIDS_DOMAIN_DIMENSION],
        expectedColumns: [
          "timestamp_rill_TIME_GRAIN_HOUR",
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        updatedRows: [AD_BIDS_PUBLISHER_DIMENSION],
        updatedColumns: [
          "timestamp_rill_TIME_GRAIN_HOUR",
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        expectedRequest: {
          dimensions: [
            { name: "publisher" },
            {
              name: "timestamp",
              timeGrain: "TIME_GRAIN_HOUR",
              timeZone: "UTC",
              alias: "Time hour",
            },
          ],
          measures: [{ name: "impressions" }],
          pivotOn: ["Time hour"],
          sort: [{ name: "publisher", desc: false }],
        },
      },

      {
        title: "Change measure",

        mutations: [AD_BIDS_OPEN_IMP_TDD, AD_BIDS_SET_DOMAIN_COMPARE_DIMENSION],
        expectedRows: [AD_BIDS_DOMAIN_DIMENSION],
        expectedColumns: [
          "timestamp_rill_TIME_GRAIN_HOUR",
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        updatedRows: [AD_BIDS_DOMAIN_DIMENSION],
        updatedColumns: [
          "timestamp_rill_TIME_GRAIN_HOUR",
          AD_BIDS_BID_PRICE_MEASURE,
        ],

        expectedRequest: {
          dimensions: [
            { name: "domain" },
            {
              name: "timestamp",
              timeGrain: "TIME_GRAIN_HOUR",
              timeZone: "UTC",
              alias: "Time hour",
            },
          ],
          measures: [{ name: "bid_price" }],
          pivotOn: ["Time hour"],
          sort: [{ name: "domain", desc: true }],
        },
      },

      {
        title: "Add measure and dimension",

        mutations: [AD_BIDS_OPEN_IMP_TDD, AD_BIDS_SET_DOMAIN_COMPARE_DIMENSION],
        expectedRows: [AD_BIDS_DOMAIN_DIMENSION],
        expectedColumns: [
          "timestamp_rill_TIME_GRAIN_HOUR",
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        updatedRows: [AD_BIDS_DOMAIN_DIMENSION, AD_BIDS_PUBLISHER_DIMENSION],
        updatedColumns: [
          "timestamp_rill_TIME_GRAIN_HOUR",
          AD_BIDS_BID_PRICE_MEASURE,
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        expectedRequest: {
          dimensions: [
            { name: "domain" },
            { name: "publisher" },
            {
              name: "timestamp",
              timeGrain: "TIME_GRAIN_HOUR",
              timeZone: "UTC",
              alias: "Time hour",
            },
          ],
          measures: [{ name: "bid_price" }, { name: "impressions" }],
          pivotOn: ["Time hour"],
          sort: [{ name: "domain", desc: true }],
        },
      },
    ];

    testCases.forEach((testCase) => {
      it(testCase.title, async () => {
        await runTest(
          testCase,
          (exploreState, timeControlState) =>
            getTDDAggregationRequest({
              metricsViewName: AD_BIDS_METRICS_NAME,
              exploreState,
              timeControlState,
              metricsViewSpec: AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
              exploreSpec: AD_BIDS_EXPLORE,
              dimensionSearchText: "",
              isScheduled: true,
            })!,
        );
      });
    });
  });

  describe("From Pivot", () => {
    const testCases: TestCase[] = [
      {
        title: "Changing dimension and measure for non-flat pivot",
        mutations: [AD_BIDS_OPEN_PIVOT_WITH_ALL_FIELDS],
        expectedRows: [
          AD_BIDS_PUBLISHER_DIMENSION,
          "timestamp_rill_TIME_GRAIN_HOUR",
        ],
        expectedColumns: [
          AD_BIDS_DOMAIN_DIMENSION,
          "timestamp_rill_TIME_GRAIN_DAY",
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        updatedRows: [
          AD_BIDS_DOMAIN_DIMENSION,
          "timestamp_rill_TIME_GRAIN_HOUR",
        ],
        updatedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          "timestamp_rill_TIME_GRAIN_DAY",
          AD_BIDS_BID_PRICE_MEASURE,
        ],

        expectedRequest: {
          dimensions: [
            { name: "domain" },
            {
              name: "timestamp",
              timeGrain: "TIME_GRAIN_HOUR",
              timeZone: "UTC",
              alias: "Time hour",
            },
            { name: "publisher" },
            {
              name: "timestamp",
              timeGrain: "TIME_GRAIN_DAY",
              timeZone: "UTC",
              alias: "Time day",
            },
          ],
          measures: [{ name: "bid_price" }],
          pivotOn: ["publisher", "Time day"],
          sort: [{ desc: false, name: "domain" }],
        },
      },

      {
        title: "Changing dimension and measure for flat pivot",
        mutations: [AD_BIDS_FLAT_PIVOT_TABLE],
        expectedRows: [],
        expectedColumns: [
          AD_BIDS_DOMAIN_DIMENSION,
          "timestamp_rill_TIME_GRAIN_DAY",
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        updatedRows: [],
        updatedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          "timestamp_rill_TIME_GRAIN_DAY",
          AD_BIDS_BID_PRICE_MEASURE,
        ],

        expectedRequest: {
          dimensions: [
            { name: "publisher" },
            {
              name: "timestamp",
              timeGrain: "TIME_GRAIN_DAY",
              timeZone: "UTC",
              alias: "Time day",
            },
          ],
          measures: [{ name: "bid_price" }],
          sort: [{ desc: true, name: "bid_price" }],
        },
      },

      {
        title: "Sort by time dimension",
        mutations: [
          AD_BIDS_FLAT_PIVOT_TABLE,
          AD_BIDS_SET_TIME_PIVOT_FILTER("timestamp_rill_TIME_GRAIN_DAY"),
        ],
        expectedRows: [],
        expectedColumns: [
          AD_BIDS_DOMAIN_DIMENSION,
          "timestamp_rill_TIME_GRAIN_DAY",
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        updatedRows: [],
        updatedColumns: [
          AD_BIDS_PUBLISHER_DIMENSION,
          "timestamp_rill_TIME_GRAIN_DAY",
          AD_BIDS_IMPRESSIONS_MEASURE,
        ],

        expectedRequest: {
          dimensions: [
            { name: "publisher" },
            {
              name: "timestamp",
              timeGrain: "TIME_GRAIN_DAY",
              timeZone: "UTC",
              alias: "Time day",
            },
          ],
          measures: [{ name: "impressions" }],
          sort: [{ desc: true, name: "Time day" }],
        },
      },
    ];

    testCases.forEach((testCase) => {
      it(testCase.title, async () => {
        await runTest(
          testCase,
          (exploreState, timeControlState) =>
            getPivotAggregationRequest({
              metricsViewName: AD_BIDS_METRICS_NAME,
              timeDimension:
                AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME.timeDimension!,
              exploreState,
              timeRange: {
                start: timeControlState.timeStart,
                end: timeControlState.timeEnd,
              },
              rows: exploreState.pivot.rows,
              columns: splitPivotChips(exploreState.pivot.columns),
              comparisonTime: timeControlState.selectedComparisonTimeRange
                ? {
                    start: timeControlState.comparisonTimeStart,
                    end: timeControlState.comparisonTimeEnd,
                  }
                : undefined,
              enableComparison: !!timeControlState.showTimeComparison,
              isFlat: exploreState.pivot.tableMode === "flat",
              pivotState: exploreState.pivot,
            })!,
        );
      });
    });
  });
});

type TestCase = {
  title: string;

  mutations: TestDashboardMutation[];

  expectedRows: string[];
  expectedColumns: string[];

  updatedRows: string[];
  updatedColumns: string[];
  expectedRequest: V1MetricsViewAggregationRequest;
};

async function runTest(
  {
    mutations,

    expectedRows,
    expectedColumns,

    updatedRows,
    updatedColumns,
    expectedRequest,
  }: TestCase,
  aggregationRequestGetter: (
    exploreState: ExploreState,
    timeControlState: TimeControlState,
  ) => V1MetricsViewAggregationRequest,
) {
  metricsExplorerStore.init(
    AD_BIDS_EXPLORE_NAME,
    getInitExploreStateForTest(
      AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
      AD_BIDS_EXPLORE,
      AD_BIDS_TIME_RANGE_SUMMARY,
    ),
  );

  await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, [
    leaderboardContextCorrection,
  ]);
  await applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

  const exploreState = get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME];
  const timeControlState = getTimeControlState(
    AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
    AD_BIDS_EXPLORE,
    AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
    exploreState,
  )!;
  timeControlState.ready = true;

  const request = aggregationRequestGetter(exploreState, timeControlState);
  const { rows, columns } = extractRowsAndColumns(request);
  expect(rows).toEqual(expectedRows);
  expect(columns).toEqual(expectedColumns);

  const newRequest = buildAggregationRequest(request, [
    aggregationRequestWithRowsAndColumns({
      exploreSpec: AD_BIDS_EXPLORE,
      rows: updatedRows,
      columns: updatedColumns,
      showTimeComparison: exploreState.showTimeComparison,
      selectedTimezone: exploreState.selectedTimezone,
    }),
  ]);
  // Remove keys that have "undefined" value. Since they are equivalent, we can skip specifying them in expected requests.
  const cleanedRequest = cleanAggregationRequestForAssertion(newRequest);
  expect(cleanedRequest).toEqual({
    ...expectedRequest,
    // Repeated fields that need not be repeated in expected requests.
    metricsView: AD_BIDS_METRICS_NAME,
    instanceId: "",
    offset: "0",
  });
}

function cleanAggregationRequestForAssertion(
  request: V1MetricsViewAggregationRequest,
) {
  const cleanedRequest = {
    ...request,
  };

  // Time ranges are not targeted in the tests.
  delete cleanedRequest.timeRange;
  delete cleanedRequest.comparisonTimeRange;

  Object.keys(cleanedRequest).forEach((key) => {
    if (cleanedRequest[key] === undefined) delete cleanedRequest[key];
  });

  return cleanedRequest;
}

// There seems to be an issue with setting context column.
// Since it is marked deprecated, all the url <=> dashboard transforms dont populate it.
// But looks like it is used extensively in the dimension table. So we need this hack to get the tests to not fail.
// TODO: either really deprecate it or account for it in transforms.
const leaderboardContextCorrection: TestDashboardMutation = (mut) => {
  mut.dashboard.leaderboardContextColumn = LeaderboardContextColumn.HIDDEN;
};
