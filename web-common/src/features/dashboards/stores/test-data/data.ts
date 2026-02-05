import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { getRillDefaultExploreState } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state.ts";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";
import { getDefaultExplorePreset } from "@rilldata/web-common/features/dashboards/url-state/getDefaultExplorePreset";
import { getLocalIANA } from "@rilldata/web-common/lib/time/timezone";
import {
  getOffset,
  getStartOfPeriod,
} from "@rilldata/web-common/lib/time/transforms";
import {
  type DashboardTimeControls,
  Period,
  TimeOffsetType,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import { DashboardState_ActivePage } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  TypeCode,
  V1ExploreComparisonMode,
  type V1ExplorePreset,
  V1ExploreSortType,
  type V1ExploreSpec,
  V1ExploreWebView,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";

export const AD_BIDS_NAME = "AdBids";
export const AD_BIDS_METRICS_NAME = AD_BIDS_NAME + "_metrics";
export const AD_BIDS_EXPLORE_NAME = AD_BIDS_NAME + "_explore";
export const AD_BIDS_SOURCE_NAME = "AdBids_Source";
export const AD_BIDS_MIRROR_NAME = "AdBids_mirror";

export const AD_BIDS_IMPRESSIONS_MEASURE = "impressions";
export const AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN = "impressions_no_grain";
export const AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN = "impressions_day_grain";
export const AD_BIDS_IMPRESSIONS_MEASURE_WINDOW = "impressions_window";
export const AD_BIDS_BID_PRICE_MEASURE = "bid_price";
export const AD_BIDS_PUBLISHER_COUNT_MEASURE = "publisher_count";
export const AD_BIDS_PUBLISHER_DIMENSION = "publisher";
export const AD_BIDS_DOMAIN_DIMENSION = "domain";
export const AD_BIDS_COUNTRY_DIMENSION = "country";
export const AD_BIDS_PUBLISHER_IS_NULL_DOMAIN = "publisher_is_null";
export const AD_BIDS_TIMESTAMP_DIMENSION = "timestamp";

export const AD_BIDS_INIT_MEASURES: MetricsViewSpecMeasure[] = [
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE,
    expression: "count(*)",
  },
  {
    name: AD_BIDS_BID_PRICE_MEASURE,
    expression: "avg(bid_price)",
  },
];
export const AD_BIDS_THREE_MEASURES: MetricsViewSpecMeasure[] = [
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE,
    expression: "count(*)",
  },
  {
    name: AD_BIDS_BID_PRICE_MEASURE,
    expression: "avg(bid_price)",
  },
  {
    name: AD_BIDS_PUBLISHER_COUNT_MEASURE,
    expression: "count_distinct(publisher)",
  },
];
export const AD_BIDS_ADVANCED_MEASURES: MetricsViewSpecMeasure[] = [
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE,
    expression: "count(*)",
  },
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE_DAY_GRAIN,
    requiredDimensions: [
      {
        name: AD_BIDS_TIMESTAMP_DIMENSION,
        timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
      },
    ],
  },
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE_NO_GRAIN,
    requiredDimensions: [
      {
        name: AD_BIDS_TIMESTAMP_DIMENSION,
        timeGrain: V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      },
    ],
  },
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE_WINDOW,
    window: {
      partition: true,
    },
  },
];
export const AD_BIDS_INIT_DIMENSIONS: MetricsViewSpecDimension[] = [
  {
    name: AD_BIDS_PUBLISHER_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: AD_BIDS_DOMAIN_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
];
export const AD_BIDS_THREE_DIMENSIONS: MetricsViewSpecDimension[] = [
  {
    name: AD_BIDS_PUBLISHER_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: AD_BIDS_DOMAIN_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: AD_BIDS_COUNTRY_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
];

const Hour = 1000 * 60 * 60;
export const TestTimeConstants = {
  NOW: new Date(),
  LAST_6_HOURS: new Date(Date.now() - Hour * 6),
  LAST_12_HOURS: new Date(Date.now() - Hour * 12),
  LAST_18_HOURS: new Date(Date.now() - Hour * 18),
  LAST_DAY: new Date(Date.now() - Hour * 24),
};
export const TestTimeOffsetConstants = {
  NOW: getOffsetByHour(TestTimeConstants.NOW),
  LAST_6_HOURS: getOffsetByHour(TestTimeConstants.LAST_6_HOURS),
  LAST_12_HOURS: getOffsetByHour(TestTimeConstants.LAST_12_HOURS),
  LAST_18_HOURS: getOffsetByHour(TestTimeConstants.LAST_18_HOURS),
  LAST_DAY: getOffsetByHour(TestTimeConstants.LAST_DAY),
};
export const AD_BIDS_DEFAULT_TIME_RANGE = {
  name: TimeRangePreset.QUARTER_TO_DATE,
  interval: V1TimeGrain.TIME_GRAIN_WEEK,
  start: TestTimeConstants.LAST_DAY,
  end: new Date(TestTimeConstants.NOW.getTime() + 1),
};
export const AD_BIDS_DEFAULT_URL_TIME_RANGE = {
  name: TimeRangePreset.ALL_TIME,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
};

export const AD_BIDS_TIME_RANGE_SUMMARY: V1MetricsViewTimeRangeResponse = {
  timeRangeSummary: {
    min: TestTimeConstants.LAST_DAY.toISOString(),
    max: TestTimeConstants.NOW.toISOString(),
  },
};

export const AD_BIDS_METRICS_INIT: V1MetricsViewSpec = {
  displayName: AD_BIDS_NAME,
  table: AD_BIDS_SOURCE_NAME,
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: AD_BIDS_INIT_DIMENSIONS,
};
export const AD_BIDS_METRICS_INIT_WITH_TIME: V1MetricsViewSpec = {
  ...AD_BIDS_METRICS_INIT,
  timeDimension: AD_BIDS_TIMESTAMP_DIMENSION,
};
export const AD_BIDS_METRICS_WITH_DELETED_MEASURE: V1MetricsViewSpec = {
  displayName: AD_BIDS_NAME,
  table: AD_BIDS_SOURCE_NAME,
  measures: [
    {
      name: AD_BIDS_IMPRESSIONS_MEASURE,
      expression: "count(*)",
    },
  ],
  dimensions: AD_BIDS_INIT_DIMENSIONS,
};
export const AD_BIDS_METRICS_WITH_BOOL_DIMENSION: V1MetricsViewSpec = {
  displayName: AD_BIDS_NAME,
  table: AD_BIDS_SOURCE_NAME,
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: [
    ...AD_BIDS_INIT_DIMENSIONS,
    {
      name: AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
      expression: "case when publisher is null then true else false end",
      dataType: { code: TypeCode.CODE_BOOL },
    },
  ],
};
export const AD_BIDS_METRICS_3_MEASURES_DIMENSIONS: V1MetricsViewSpec = {
  displayName: AD_BIDS_NAME,
  table: AD_BIDS_SOURCE_NAME,
  measures: AD_BIDS_THREE_MEASURES,
  dimensions: AD_BIDS_THREE_DIMENSIONS,
  timeDimension: AD_BIDS_TIMESTAMP_DIMENSION,
};
export const AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME: V1MetricsViewSpec =
  {
    ...AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
    timeDimension: AD_BIDS_TIMESTAMP_DIMENSION,
  };

export const AD_BIDS_EXPLORE_INIT: V1ExploreSpec = {
  displayName: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_METRICS_NAME,
  measures: AD_BIDS_INIT_MEASURES.map((m) => m.name!),
  dimensions: AD_BIDS_INIT_DIMENSIONS.map((d) => d.name!),
};
export const AD_BIDS_EXPLORE_WITH_DELETED_MEASURE: V1ExploreSpec = {
  displayName: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: [AD_BIDS_IMPRESSIONS_MEASURE],
  dimensions: AD_BIDS_INIT_DIMENSIONS.map((d) => d.name!),
};
export const AD_BIDS_EXPLORE_WITH_THREE_MEASURES: V1ExploreSpec = {
  displayName: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: AD_BIDS_THREE_MEASURES.map((m) => m.name!),
  dimensions: AD_BIDS_INIT_DIMENSIONS.map((d) => d.name!),
};
export const AD_BIDS_EXPLORE_WITH_DELETED_DIMENSION: V1ExploreSpec = {
  displayName: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: AD_BIDS_INIT_MEASURES.map((m) => m.name!),
  dimensions: [AD_BIDS_PUBLISHER_DIMENSION],
};
export const AD_BIDS_EXPLORE_WITH_THREE_DIMENSIONS: V1ExploreSpec = {
  displayName: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: AD_BIDS_INIT_MEASURES.map((m) => m.name!),
  dimensions: AD_BIDS_THREE_DIMENSIONS.map((d) => d.name!),
};
export const AD_BIDS_EXPLORE_WITH_BOOL_DIMENSION: V1ExploreSpec = {
  displayName: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: AD_BIDS_INIT_MEASURES.map((m) => m.name!),
  dimensions: [
    ...AD_BIDS_INIT_DIMENSIONS.map((d) => d.name!),
    AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
  ],
};
export const AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS: V1ExploreSpec = {
  displayName: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_METRICS_NAME,
  measures: AD_BIDS_THREE_MEASURES.map((m) => m.name!),
  dimensions: AD_BIDS_THREE_DIMENSIONS.map((d) => d.name!),
};

export const AD_BIDS_PRESET: V1ExplorePreset = {
  timeRange: "P7D",
  timezone: "Asia/Kathmandu",
  compareTimeRange: "rill-PP",
  comparisonMode: V1ExploreComparisonMode.EXPLORE_COMPARISON_MODE_TIME,
  measures: [AD_BIDS_IMPRESSIONS_MEASURE],
  dimensions: [AD_BIDS_PUBLISHER_DIMENSION],
  exploreSortBy: AD_BIDS_IMPRESSIONS_MEASURE,
  exploreSortAsc: true,
  exploreSortType: V1ExploreSortType.EXPLORE_SORT_TYPE_PERCENT,
};
export const AD_BIDS_PRESET_WITHOUT_TIMESTAMP: V1ExplorePreset = {
  measures: [AD_BIDS_IMPRESSIONS_MEASURE],
  dimensions: [AD_BIDS_PUBLISHER_DIMENSION],
  exploreSortBy: AD_BIDS_IMPRESSIONS_MEASURE,
  exploreSortAsc: true,
  exploreSortType: V1ExploreSortType.EXPLORE_SORT_TYPE_PERCENT,
};
export const AD_BIDS_DIMENSION_TABLE_PRESET: V1ExplorePreset = {
  exploreExpandedDimension: AD_BIDS_DOMAIN_DIMENSION,
};
export const AD_BIDS_TIME_DIMENSION_DETAILS_PRESET: V1ExplorePreset = {
  view: V1ExploreWebView.EXPLORE_WEB_VIEW_TIME_DIMENSION,
  timeDimensionMeasure: AD_BIDS_IMPRESSIONS_MEASURE,
  timeDimensionChartType: "stacked_bar",
};
export const AD_BIDS_PIVOT_PRESET: V1ExplorePreset = {
  view: V1ExploreWebView.EXPLORE_WEB_VIEW_PIVOT,
  pivotRows: ["publisher", "time.hour"],
  pivotCols: ["domain", "time.day", "impressions"],
  pivotSortBy: "time.day",
  pivotTableMode: "nest",
  pivotSortAsc: true,
};

export const AD_BIDS_BASE_PRESET = getDefaultExplorePreset(
  AD_BIDS_EXPLORE_INIT,
  AD_BIDS_METRICS_INIT,
  undefined,
);

export function getOffsetByHour(time: Date) {
  return getOffset(
    getStartOfPeriod(time, Period.HOUR, getLocalIANA()),
    Period.HOUR,
    TimeOffsetType.ADD,
    getLocalIANA(),
  );
}

export const AD_BIDS_BASE_FILTER = createAndExpression([
  createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google", "Facebook"]),
  createInExpression(AD_BIDS_DOMAIN_DIMENSION, ["google.com"]),
]);

export const AD_BIDS_EXCLUDE_FILTER = createAndExpression([
  createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google", "Facebook"], true),
  createInExpression(AD_BIDS_DOMAIN_DIMENSION, ["google.com"]),
]);

export const AD_BIDS_CLEARED_FILTER = createAndExpression([
  createInExpression(AD_BIDS_PUBLISHER_DIMENSION, ["Google", "Facebook"], true),
]);

// parsed time controls won't have start & end
export const ALL_TIME_PARSED_TEST_CONTROLS = {
  name: TimeRangePreset.QUARTER_TO_DATE,
  interval: V1TimeGrain.TIME_GRAIN_WEEK,
} as DashboardTimeControls;

export const LAST_6_HOURS_TEST_CONTROLS = {
  name: TimeRangePreset.LAST_SIX_HOURS,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
  start: TestTimeConstants.LAST_6_HOURS,
  end: TestTimeConstants.NOW,
} as DashboardTimeControls;

// parsed time controls won't have start & end
export const LAST_6_HOURS_TEST_PARSED_CONTROLS = {
  name: TimeRangePreset.LAST_SIX_HOURS,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
} as DashboardTimeControls;

export const CUSTOM_TEST_CONTROLS = {
  name: TimeRangePreset.CUSTOM,
  interval: V1TimeGrain.TIME_GRAIN_MINUTE,
  start: TestTimeConstants.LAST_18_HOURS,
  end: TestTimeConstants.LAST_12_HOURS,
} as DashboardTimeControls;

export const AD_BIDS_PIVOT_ENTITY: Partial<ExploreState> = {
  activePage: DashboardState_ActivePage.PIVOT,
  pivot: {
    rows: [
      {
        id: AD_BIDS_PUBLISHER_DIMENSION,
        type: PivotChipType.Dimension,
        title: AD_BIDS_PUBLISHER_DIMENSION,
      },
      {
        id: V1TimeGrain.TIME_GRAIN_HOUR,
        type: PivotChipType.Time,
        title: "hour",
      },
    ],

    columns: [
      {
        id: AD_BIDS_DOMAIN_DIMENSION,
        type: PivotChipType.Dimension,
        title: AD_BIDS_DOMAIN_DIMENSION,
      },
      {
        id: V1TimeGrain.TIME_GRAIN_DAY,
        type: PivotChipType.Time,
        title: "day",
      },
      {
        id: AD_BIDS_IMPRESSIONS_MEASURE,
        type: PivotChipType.Measure,
        title: AD_BIDS_IMPRESSIONS_MEASURE,
      },
    ],
    expanded: {},
    sorting: [],
    columnPage: 1,
    rowPage: 1,
    enableComparison: true,
    activeCell: null,
    tableMode: "nest",
  },
};

export const AD_BIDS_RILL_DEFAULT_EXPLORE_STATE = getRillDefaultExploreState(
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
);
export const AD_BIDS_RILL_DEFAULT_EXPLORE_URL_PARAMS =
  getRillDefaultExploreUrlParams(
    AD_BIDS_METRICS_3_MEASURES_DIMENSIONS_WITH_TIME,
    AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
    AD_BIDS_TIME_RANGE_SUMMARY.timeRangeSummary,
  );
