import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  TypeCode,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

export const PIVOT_TEST_NAME = "PivotTest";
export const PIVOT_TEST_METRICS_NAME = "mv1";
export const PIVOT_TEST_SOURCE_NAME = "PivotTest_Source";

export const PIVOT_TOTAL_MEASURE = "total";
export const PIVOT_REVENUE_MEASURE = "revenue";
export const PIVOT_OTHER_MEASURE = "other_measure";

// Row dimensions.
export const PIVOT_COUNTRY_DIMENSION = "country";
export const PIVOT_CITY_DIMENSION = "city";
export const PIVOT_OUTER_DIMENSION = "outer";
export const PIVOT_INNER_DIMENSION = "inner";
export const PIVOT_BOROUGH_DIMENSION = "borough";

// Column dimensions.
export const PIVOT_REGION_DIMENSION = "region";
export const PIVOT_CATEGORY_DIMENSION = "category";
export const PIVOT_PRODUCT_DIMENSION = "product";
export const PIVOT_STATUS_DIMENSION = "status";
export const PIVOT_TYPE_DIMENSION = "type";
export const PIVOT_QUARTER_DIMENSION = "quarter";
export const PIVOT_ENV_DIMENSION = "env";

export const PIVOT_MEASURES: MetricsViewSpecMeasure[] = [
  {
    name: PIVOT_TOTAL_MEASURE,
    expression: "count(*)",
  },
  {
    name: PIVOT_REVENUE_MEASURE,
    expression: "sum(revenue)",
  },
  {
    name: PIVOT_OTHER_MEASURE,
    expression: "avg(revenue)",
  },
];

export const PIVOT_DIMENSIONS: MetricsViewSpecDimension[] = [
  {
    name: PIVOT_COUNTRY_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_CITY_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_OUTER_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_INNER_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_BOROUGH_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_REGION_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_CATEGORY_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_PRODUCT_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_STATUS_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_TYPE_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_QUARTER_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
  {
    name: PIVOT_ENV_DIMENSION,
    dataType: { code: TypeCode.CODE_STRING },
  },
];

/**
 * Metrics view covering every measure and dimension referenced by the pivot
 * click-to-filter tests. It has no time dimension, so MetricsViewsProvider
 * reports it ready as soon as the spec lands, without a time range query.
 */
export const PIVOT_METRICS_INIT: V1MetricsViewSpec = {
  displayName: PIVOT_TEST_NAME,
  table: PIVOT_TEST_SOURCE_NAME,
  measures: PIVOT_MEASURES,
  dimensions: PIVOT_DIMENSIONS,
};
