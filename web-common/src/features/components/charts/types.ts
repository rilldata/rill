import type {
  CartesianChartProvider,
  CartesianChartSpec,
  CircularChartProvider,
  CircularChartSpec,
  ComboChartProvider,
  ComboChartSpec,
  FunnelChartProvider,
  FunnelChartSpec,
  HeatmapChartProvider,
  HeatmapChartSpec,
} from "@rilldata/web-common/features/components/charts";
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
import type { Color } from "chroma-js";
import type { TimeUnit } from "vega-lite/build/src/timeunit";
import type { ColorScheme } from "vega-typings";

export type ChartProvider =
  | CartesianChartProvider
  | CircularChartProvider
  | ComboChartProvider
  | FunnelChartProvider
  | HeatmapChartProvider;

export type ChartSpecBase =
  | CartesianChartSpec
  | CircularChartSpec
  | FunnelChartSpec
  | HeatmapChartSpec
  | ComboChartSpec;

export type ChartSpec = ChartSpecBase & {
  vl_config?: string;
};

interface TimeRange {
  time_range: {
    start: string;
    end: string;
  };
}

export type ChartSpecAI =
  | { chart_type: "bar_chart"; spec: CartesianChartSpec & TimeRange }
  | { chart_type: "line_chart"; spec: CartesianChartSpec & TimeRange }
  | { chart_type: "area_chart"; spec: CartesianChartSpec & TimeRange }
  | { chart_type: "stacked_bar"; spec: CartesianChartSpec & TimeRange }
  | {
      chart_type: "stacked_bar_normalized";
      spec: CartesianChartSpec & TimeRange;
    }
  | { chart_type: "donut_chart"; spec: CircularChartSpec & TimeRange }
  | { chart_type: "pie_chart"; spec: CircularChartSpec & TimeRange }
  | { chart_type: "funnel_chart"; spec: FunnelChartSpec & TimeRange }
  | { chart_type: "heatmap"; spec: HeatmapChartSpec & TimeRange }
  | { chart_type: "combo_chart"; spec: ComboChartSpec & TimeRange };

export type ChartType =
  | "bar_chart"
  | "line_chart"
  | "area_chart"
  | "stacked_bar"
  | "stacked_bar_normalized"
  | "donut_chart"
  | "pie_chart"
  | "heatmap"
  | "funnel_chart"
  | "combo_chart";

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
  theme: { primary: Color; secondary: Color };
  domainValues?: ChartDomainValues;
  isDarkMode: boolean;
};

export interface ChartDomainValues {
  [key: string]: string[] | number[] | undefined;
}

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
  | "measure"
  | "-measure"
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
  colorMapping?: ColorMapping;
}

interface MarkFieldConfig {
  mark?: "bar" | "line";
}

interface TimeFieldConfig {
  timeUnit?: string; // For temporal fields
}

interface QuantitativeFieldConfig {
  zeroBasedOrigin?: boolean; // Default is false
  min?: number;
  max?: number;
  showTotal?: boolean;
  colorRange?: ColorRangeMapping;
}

interface BaseFieldConfig {
  field: string;
  type: "quantitative" | "ordinal" | "nominal" | "temporal" | "value";
  showAxisTitle?: boolean; // Default is false
  fields?: string[]; // To support multi metric chart variants
}

export type FieldConfig<
  TInclude extends "nominal" | "quantitative" | "time" | "mark" =
    | "nominal"
    | "quantitative"
    | "time"
    | "mark",
> = BaseFieldConfig &
  ("nominal" extends TInclude ? NominalFieldConfig : object) &
  ("quantitative" extends TInclude ? QuantitativeFieldConfig : object) &
  ("time" extends TInclude ? TimeFieldConfig : object) &
  ("mark" extends TInclude ? MarkFieldConfig : object);

export interface CommonChartProperties {
  metrics_view: string;
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

export type ColorMapping = { value: string; color: string }[];

export type ColorRangeMapping =
  | {
      mode: "scheme";
      scheme: ColorScheme | "sequential" | "diverging";
    }
  | {
      mode: "gradient";
      start: string;
      end: string;
    };
