import type {
  V1Expression,
  V1MetricsViewAggregationDimension,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationSort,
} from "@rilldata/web-common/runtime-client";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1MetricsViewAggregationResponseDataItem,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { ComponentType, SvelteComponent } from "svelte";
import { type Readable } from "svelte/store";
import type { ChartType } from "./";

export type ChartDataQuery = Readable<{
  isFetching: boolean;
  error: HTTPError | null;
  data: V1MetricsViewAggregationResponseDataItem[] | undefined;
}>;

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
  timeUnit?: string;
  format?: string;
}

export type ChartSortDirection = "x" | "y" | "-x" | "-y";

export interface FieldConfig {
  field: string;
  showAxisTitle?: boolean; // Default is false
  zeroBasedOrigin?: boolean; // Default is false
  type: "quantitative" | "ordinal" | "nominal" | "temporal";
  timeUnit?: string; // For temporal fields
  sort?: ChartSortDirection;
  limit?: number;
  showNull?: boolean;
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

export interface ChartMetadata {
  type: ChartType;
  icon: ComponentType<SvelteComponent>;
  title: string;
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
