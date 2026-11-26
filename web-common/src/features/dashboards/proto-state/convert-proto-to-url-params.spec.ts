import {
  protoBase64,
  Timestamp,
  type PartialMessage,
} from "@bufbuild/protobuf";
import { convertProtoToUrlParams } from "@rilldata/web-common/features/dashboards/proto-state/convert-proto-to-url-params";
import { toExpressionProto } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import {
  createAndExpression,
  createBinaryExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  AD_BIDS_BID_PRICE_MEASURE,
  AD_BIDS_DOMAIN_DIMENSION,
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_IMPRESSIONS_MEASURE,
  AD_BIDS_METRICS_INIT_WITH_TIME,
  AD_BIDS_PUBLISHER_DIMENSION,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data";
import { convertURLSearchParamsToExploreState } from "@rilldata/web-common/features/dashboards/url-state/convertURLSearchParamsToExploreState";
import { TimeGrain } from "@rilldata/web-common/proto/gen/rill/runtime/v1/time_grain_pb";
import {
  DashboardState,
  DashboardState_ActivePage,
  DashboardState_LeaderboardSortDirection,
  DashboardState_LeaderboardSortType,
  DashboardState_PivotTableMode,
  DashboardState_PivotTotalsRowPosition,
} from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import { V1Operation } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

// The legacy encoder stored milliseconds in `seconds`.
const Start = new Timestamp({
  seconds: BigInt(new Date("2024-01-01T00:00:00Z").getTime()),
});
const End = new Timestamp({
  seconds: BigInt(new Date("2024-02-01T00:00:00Z").getTime()),
});

const TestCases: {
  title: string;
  state: PartialMessage<DashboardState>;
  expected: Record<string, string>;
}[] = [
  {
    title: "filters with legacy measure filters",
    state: {
      where: toExpressionProto(
        createAndExpression([
          createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
        ]),
      ),
      having: [
        {
          name: AD_BIDS_DOMAIN_DIMENSION,
          filter: toExpressionProto(
            createBinaryExpression(
              AD_BIDS_IMPRESSIONS_MEASURE,
              V1Operation.OPERATION_GT,
              10,
            ),
          ),
        },
      ],
    },
    expected: {
      f: "publisher IN ('Google') AND domain having (impressions GT 10)",
    },
  },
  {
    title: "legacy measure filters with a filter that is not a joiner",
    state: {
      where: toExpressionProto(
        createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google"]),
      ),
      having: [
        {
          name: AD_BIDS_DOMAIN_DIMENSION,
          filter: toExpressionProto(
            createBinaryExpression(
              AD_BIDS_IMPRESSIONS_MEASURE,
              V1Operation.OPERATION_GT,
              10,
            ),
          ),
        },
      ],
    },
    expected: {
      f: "publisher IN ('Google') AND domain having (impressions GT 10)",
    },
  },
  {
    title: "custom time ranges",
    state: {
      timeRange: { name: "CUSTOM", timeStart: Start, timeEnd: End },
      compareTimeRange: {
        name: "CUSTOM_COMPARISON_RANGE",
        timeStart: Start,
        timeEnd: End,
      },
      showTimeComparison: true,
      timeGrain: TimeGrain.TIME_GRAIN_DAY,
      selectedTimezone: "Asia/Kathmandu",
    },
    expected: {
      tr: "2024-01-01T00:00:00.000Z,2024-02-01T00:00:00.000Z",
      compare_tr: "2024-01-01T00:00:00.000Z,2024-02-01T00:00:00.000Z",
      grain: "day",
      tz: "Asia/Kathmandu",
    },
  },
  {
    title: "legacy time range names",
    state: {
      timeRange: { name: "LAST_7_DAYS" },
      compareTimeRange: { name: "P1W" },
      showTimeComparison: true,
    },
    expected: { tr: "P7D", compare_tr: "rill-PW" },
  },
  {
    title: "time comparison turned off",
    state: {
      timeRange: { name: "P7D" },
      compareTimeRange: { name: "rill-PW" },
      showTimeComparison: false,
    },
    expected: { tr: "P7D", compare_tr: "" },
  },
  {
    title: "time comparison turned on without a comparison range",
    state: {
      timeRange: { name: "P7D" },
      showTimeComparison: true,
    },
    expected: { tr: "P7D", compare_tr: "rill-PP" },
  },
  {
    title: "time grain without a time range",
    state: { timeGrain: TimeGrain.TIME_GRAIN_DAY },
    expected: {},
  },
  {
    title: "explore view",
    state: {
      activePage: DashboardState_ActivePage.DEFAULT,
      allMeasuresVisible: true,
      visibleDimensions: [AD_BIDS_PUBLISHER_DIMENSION],
      leaderboardMeasure: AD_BIDS_BID_PRICE_MEASURE,
      leaderboardSortDirection:
        DashboardState_LeaderboardSortDirection.ASCENDING,
      leaderboardSortType: DashboardState_LeaderboardSortType.DELTA_PERCENT,
      leaderboardMeasures: [AD_BIDS_BID_PRICE_MEASURE],
    },
    expected: {
      view: "explore",
      measures: "*",
      dims: AD_BIDS_PUBLISHER_DIMENSION,
      sort_by: AD_BIDS_BID_PRICE_MEASURE,
      sort_dir: "ASC",
      sort_type: "delta_percent",
      leaderboard_measures: AD_BIDS_BID_PRICE_MEASURE,
      compare_dim: "",
      expand_dim: "",
    },
  },
  {
    title: "dimension table",
    state: {
      activePage: DashboardState_ActivePage.DIMENSION_TABLE,
      selectedDimension: AD_BIDS_PUBLISHER_DIMENSION,
    },
    expected: {
      view: "explore",
      compare_dim: "",
      expand_dim: AD_BIDS_PUBLISHER_DIMENSION,
    },
  },
  {
    title: "time dimension details",
    state: {
      activePage: DashboardState_ActivePage.TIME_DIMENSIONAL_DETAIL,
      expandedMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
      chartType: "grouped_bar",
      comparisonDimension: AD_BIDS_DOMAIN_DIMENSION,
    },
    expected: {
      view: "tdd",
      measure: AD_BIDS_IMPRESSIONS_MEASURE,
      chart_type: "bar",
      compare_dim: AD_BIDS_DOMAIN_DIMENSION,
      expand_dim: "",
    },
  },
  {
    title: "pivot",
    state: {
      activePage: DashboardState_ActivePage.PIVOT,
      pivotRowAllDimensions: [
        {
          element: {
            case: "pivotTimeDimension",
            value: TimeGrain.TIME_GRAIN_DAY,
          },
        },
        {
          element: {
            case: "pivotDimension",
            value: AD_BIDS_PUBLISHER_DIMENSION,
          },
        },
      ],
      pivotColumnAllDimensions: [
        {
          element: { case: "pivotDimension", value: AD_BIDS_DOMAIN_DIMENSION },
        },
      ],
      pivotColumnMeasures: [AD_BIDS_IMPRESSIONS_MEASURE],
      pivotSort: [{ id: "timestamp_rill_TIME_GRAIN_DAY", desc: true }],
      pivotTableMode: DashboardState_PivotTableMode.NEST,
      pivotShowTotalsRow: false,
      pivotTotalsRowPosition: DashboardState_PivotTotalsRowPosition.BOTTOM,
    },
    expected: {
      view: "pivot",
      rows: `time.day,${AD_BIDS_PUBLISHER_DIMENSION}`,
      cols: `${AD_BIDS_DOMAIN_DIMENSION},${AD_BIDS_IMPRESSIONS_MEASURE}`,
      sort_by: "time.day",
      sort_dir: "DESC",
      table_mode: "nest",
      show_totals_row: "false",
      totals_row_position: "bottom",
      compare_dim: "",
    },
  },
];

describe("convertProtoToUrlParams", () => {
  for (const { title, state, expected } of TestCases) {
    it(title, () => {
      const protoState = protoBase64.enc(new DashboardState(state).toBinary());
      const urlParams = convertProtoToUrlParams(protoState);

      expect(Object.fromEntries(urlParams.entries())).toEqual(expected);

      // The explore parses the params without errors.
      const { errors } = convertURLSearchParamsToExploreState(
        urlParams,
        AD_BIDS_METRICS_INIT_WITH_TIME,
        AD_BIDS_EXPLORE_INIT,
        {},
      );
      expect(errors).toEqual([]);
    });
  }
});
