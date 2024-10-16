import {
  createAndExpression,
  createInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
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
import type {
  MetricsViewSpecDimensionV2,
  MetricsViewSpecMeasureV2,
  V1ExploreSpec,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
  V1StructType,
} from "@rilldata/web-common/runtime-client";
import { TypeCode, V1TimeGrain } from "@rilldata/web-common/runtime-client";

export const AD_BIDS_NAME = "AdBids";
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

export const AD_BIDS_INIT_MEASURES: MetricsViewSpecMeasureV2[] = [
  {
    name: AD_BIDS_IMPRESSIONS_MEASURE,
    expression: "count(*)",
  },
  {
    name: AD_BIDS_BID_PRICE_MEASURE,
    expression: "avg(bid_price)",
  },
];
export const AD_BIDS_THREE_MEASURES: MetricsViewSpecMeasureV2[] = [
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
export const AD_BIDS_ADVANCED_MEASURES: MetricsViewSpecMeasureV2[] = [
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
export const AD_BIDS_INIT_DIMENSIONS: MetricsViewSpecDimensionV2[] = [
  {
    name: AD_BIDS_PUBLISHER_DIMENSION,
  },
  {
    name: AD_BIDS_DOMAIN_DIMENSION,
  },
];
export const AD_BIDS_THREE_DIMENSIONS: MetricsViewSpecDimensionV2[] = [
  {
    name: AD_BIDS_PUBLISHER_DIMENSION,
  },
  {
    name: AD_BIDS_DOMAIN_DIMENSION,
  },
  {
    name: AD_BIDS_COUNTRY_DIMENSION,
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
  name: TimeRangePreset.ALL_TIME,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
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
    interval: V1TimeGrain.TIME_GRAIN_MINUTE as any,
  },
};

export const AD_BIDS_METRICS_INIT: V1MetricsViewSpec = {
  title: AD_BIDS_NAME,
  table: AD_BIDS_SOURCE_NAME,
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: AD_BIDS_INIT_DIMENSIONS,
};
export const AD_BIDS_METRICS_INIT_WITH_TIME: V1MetricsViewSpec = {
  ...AD_BIDS_METRICS_INIT,
  timeDimension: AD_BIDS_TIMESTAMP_DIMENSION,
};
export const AD_BIDS_METRICS_WITH_DELETED_MEASURE: V1MetricsViewSpec = {
  title: AD_BIDS_NAME,
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
  title: AD_BIDS_NAME,
  table: AD_BIDS_SOURCE_NAME,
  measures: AD_BIDS_INIT_MEASURES,
  dimensions: [
    ...AD_BIDS_INIT_DIMENSIONS,
    {
      name: AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
      expression: "case when publisher is null then true else false end",
    },
  ],
};

export const AD_BIDS_EXPLORE_INIT: V1ExploreSpec = {
  title: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: AD_BIDS_INIT_MEASURES.map((m) => m.name!),
  dimensions: AD_BIDS_INIT_DIMENSIONS.map((d) => d.name!),
};
export const AD_BIDS_EXPLORE_WITH_DELETED_MEASURE: V1ExploreSpec = {
  title: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: [AD_BIDS_IMPRESSIONS_MEASURE],
  dimensions: AD_BIDS_INIT_DIMENSIONS.map((d) => d.name!),
};
export const AD_BIDS_EXPLORE_WITH_THREE_MEASURES: V1ExploreSpec = {
  title: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: AD_BIDS_THREE_MEASURES.map((m) => m.name!),
  dimensions: AD_BIDS_INIT_DIMENSIONS.map((d) => d.name!),
};
export const AD_BIDS_EXPLORE_WITH_DELETED_DIMENSION: V1ExploreSpec = {
  title: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: AD_BIDS_INIT_MEASURES.map((m) => m.name!),
  dimensions: [AD_BIDS_PUBLISHER_DIMENSION],
};
export const AD_BIDS_EXPLORE_WITH_THREE_DIMENSIONS: V1ExploreSpec = {
  title: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: AD_BIDS_INIT_MEASURES.map((m) => m.name!),
  dimensions: AD_BIDS_THREE_DIMENSIONS.map((d) => d.name!),
};
export const AD_BIDS_EXPLORE_WITH_BOOL_DIMENSION: V1ExploreSpec = {
  title: AD_BIDS_EXPLORE_NAME,
  metricsView: AD_BIDS_NAME,
  measures: AD_BIDS_INIT_MEASURES.map((m) => m.name!),
  dimensions: [
    ...AD_BIDS_INIT_DIMENSIONS.map((d) => d.name!),
    AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
  ],
};

export const AD_BIDS_SCHEMA: V1StructType = {
  fields: [
    {
      name: AD_BIDS_PUBLISHER_DIMENSION,
      type: {
        code: TypeCode.CODE_STRING,
      },
    },
    {
      name: AD_BIDS_DOMAIN_DIMENSION,
      type: {
        code: TypeCode.CODE_STRING,
      },
    },
    {
      name: AD_BIDS_COUNTRY_DIMENSION,
      type: {
        code: TypeCode.CODE_STRING,
      },
    },
    {
      name: AD_BIDS_PUBLISHER_IS_NULL_DOMAIN,
      type: {
        code: TypeCode.CODE_BOOL,
      },
    },
    {
      name: AD_BIDS_IMPRESSIONS_MEASURE,
      type: {
        code: TypeCode.CODE_INT64,
      },
    },
    {
      name: AD_BIDS_BID_PRICE_MEASURE,
      type: {
        code: TypeCode.CODE_FLOAT64,
      },
    },
  ],
};

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
  name: TimeRangePreset.ALL_TIME,
  interval: V1TimeGrain.TIME_GRAIN_HOUR,
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
