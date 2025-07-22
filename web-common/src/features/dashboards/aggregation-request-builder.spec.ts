import {
  aggregationRequestWithRowsAndColumns,
  buildAggregationRequest,
} from "@rilldata/web-common/features/dashboards/aggregation-request-builder.ts";
import { getDimensionTableAggregationRequestForTime } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-export.ts";
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
  AD_BIDS_SORT_ASC_BY_BID_PRICE,
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
import { describe, it, expect } from "vitest";
import { getTDDAggregationRequest } from "web-common/src/features/dashboards/time-dimension-details/tdd-export.ts";
import { getPivotAggregationRequest } from "web-common/src/features/dashboards/pivot/pivot-export.ts";

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
    ];

    testCases.forEach((testCase) => {
      it(testCase.title, () => {
        runTest(testCase, (exploreState, timeControlState) =>
          getDimensionTableAggregationRequestForTime({
            metricsViewName: AD_BIDS_METRICS_NAME,
            exploreState,
            timeRange: {
              start: timeControlState.timeStart,
              end: timeControlState.timeEnd,
            },
            comparisonTimeRange: undefined,
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
      it(testCase.title, () => {
        runTest(
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
    ];

    testCases.forEach((testCase) => {
      it(testCase.title, () => {
        runTest(
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
              comparisonTime: undefined,
              enableComparison: false,
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

function runTest(
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

  applyMutationsToDashboard(AD_BIDS_EXPLORE_NAME, mutations);

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
      showTimeComparison: false,
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

  delete cleanedRequest.timeRange; // Time range is not targeted in the tests.

  Object.keys(cleanedRequest).forEach((key) => {
    if (cleanedRequest[key] === undefined) delete cleanedRequest[key];
  });

  return cleanedRequest;
}
