import type {
  V1Expression,
  V1MetricsViewAggregationDimension,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationResponse,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import type { TimeUnit } from "vega-lite/build/src/timeunit";

export type ChartType =
  | "bar_chart"
  | "line_chart"
  | "area_chart"
  | "stacked_bar"
  | "stacked_bar_normalized"
  | "donut_chart"
  | "pie_chart"
  | "heatmap";

export type ChartDataQuery = CreateQueryResult<
  V1MetricsViewAggregationResponse,
  HTTPError
>;

export type ChartFieldsMap = Record<
  string,
  | MetricsViewSpecMeasure
  | MetricsViewSpecDimension
  | TimeDimensionDefinition
  | undefined
>;
export type ChartDataResult = {
  data: V1MetricsViewAggregationResponseDataItem[];
  isFetching: boolean;
  fields: ChartFieldsMap;
  error?: HTTPError | null;
};

export interface TimeDimensionDefinition {
  field: string;
  displayName: string;
  timeUnit?: TimeUnit;
  format?: string;
}

export type ChartSortDirectionOptions =
  | "x"
  | "y"
  | "-x"
  | "-y"
  | "color"
  | "-color"
  | "custom";

export type ChartSortDirection =
  | Exclude<ChartSortDirectionOptions, "custom">
  | string[];

export type ChartLegend = "none" | "top" | "bottom" | "left" | "right";

interface NominalFieldConfig {
  sort?: ChartSortDirection;
  limit?: number;
  showNull?: boolean;
  labelAngle?: number;
  legendOrientation?: ChartLegend;
}

interface QuantitativeFieldConfig {
  zeroBasedOrigin?: boolean; // Default is false
  min?: number;
  max?: number;
}

export interface FieldConfig
  extends NominalFieldConfig,
    QuantitativeFieldConfig {
  field: string;
  type: "quantitative" | "ordinal" | "nominal" | "temporal";
  showAxisTitle?: boolean; // Default is false
  timeUnit?: string; // For temporal fields
}

export interface CommonChartProperties {
  metrics_view: string;
  tooltip?: FieldConfig;
  vl_config?: string;
}

// TODO: Remove this once we have a better way to handle chart config
export interface ChartConfig extends CommonChartProperties {
  x?: FieldConfig;
  y?: FieldConfig;
  color?: FieldConfig | string;
  tooltip?: FieldConfig;
  vl_config?: string;
}

/** Temporary solution for the lack of vega lite type exports */
export interface TooltipValue {
  title?: string;
  field: string;
  format?: string;
  formatType?: string;
  type: "quantitative" | "ordinal" | "temporal" | "nominal";
}

export interface ChartQueryConfig {
  measures: V1MetricsViewAggregationMeasure[];
  dimensions: V1MetricsViewAggregationDimension[];
  sort?: V1MetricsViewAggregationSort[];
  where?: V1Expression;
  limit?: string;
}
