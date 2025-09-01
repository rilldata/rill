import type { CartesianChartSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/CartesianChart";
import type { CircularChartSpec } from "@rilldata/web-common/features/canvas/components/charts/circular-charts/CircularChart";
import type { KPIGridSpec } from "@rilldata/web-common/features/canvas/components/kpi-grid";
import type { ChartType } from "./charts/types";
import type { ImageSpec } from "./image";
import type { KPISpec } from "./kpi";
import type { LeaderboardSpec } from "./leaderboard";
import type { MarkdownSpec } from "./markdown";
import type { PivotSpec, TableSpec } from "./pivot";

export type ComponentWithMetricsView =
  | CartesianChartSpec
  | CircularChartSpec
  | PivotSpec
  | TableSpec
  | KPISpec
  | KPIGridSpec
  | LeaderboardSpec;

export type ComponentSpec = ComponentWithMetricsView | ImageSpec | MarkdownSpec;

export interface ComponentCommonProperties {
  title?: string;
  description?: string;
}

export type VeriticalAlignment = "top" | "middle" | "bottom";
export type HoritzontalAlignment = "left" | "center" | "right";
export interface ComponentAlignment {
  vertical: VeriticalAlignment;
  horizontal: HoritzontalAlignment;
}

export type ComponentComparisonOptions =
  | "previous"
  | "delta"
  | "percent_change";

export interface ComponentFilterProperties {
  time_filters?: string;
  dimension_filters?: string;
}

export interface ComponentSize {
  width: number;
  height: number;
}

export type CanvasComponentType =
  | ChartType
  | "markdown"
  | "kpi_grid"
  | "image"
  | "pivot"
  | "table"
  | "leaderboard";

interface LineChart {
  line_chart: CartesianChartSpec;
}

interface AreaChart {
  area_chart: CartesianChartSpec;
}

interface BarChart {
  bar_chart: CartesianChartSpec;
}

export type ChartTemplates = LineChart | BarChart | AreaChart;
export interface KPITemplateT {
  kpi: KPISpec;
}
export interface MarkdownTemplateT {
  markdown: MarkdownSpec;
}
export interface ImageTemplateT {
  image: ImageSpec;
}

export interface PivotTemplateT {
  pivot: PivotSpec;
}
export interface TableTemplateT {
  table: TableSpec;
}

export type TemplateSpec =
  | ChartTemplates
  | KPITemplateT
  | PivotTemplateT
  | MarkdownTemplateT
  | ImageTemplateT
  | TableTemplateT;
