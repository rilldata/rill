import type { CartesianCanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts/variants/CartesianChart";
import type { CircularCanvasChartSpec } from "@rilldata/web-common/features/canvas/components/charts/variants/CircularChart";
import type { KPIGridSpec } from "@rilldata/web-common/features/canvas/components/kpi-grid";
import type { ChartType } from "../../components/charts/types";
import type { ImageSpec } from "./image";
import type { KPISpec } from "./kpi";
import type { LeaderboardSpec } from "./leaderboard";
import type { MarkdownSpec } from "./markdown";
import type { PivotSpec, TableSpec } from "./pivot";

export type ComponentWithMetricsView =
  | CartesianCanvasChartSpec
  | CircularCanvasChartSpec
  | PivotSpec
  | TableSpec
  | KPISpec
  | KPIGridSpec
  | LeaderboardSpec;

export type ComponentSpec = ComponentWithMetricsView | ImageSpec | MarkdownSpec;

export interface ComponentCommonProperties {
  title?: string;
  description?: string;
  show_description_as_tooltip?: boolean;
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
  line_chart: CartesianCanvasChartSpec;
}

interface AreaChart {
  area_chart: CartesianCanvasChartSpec;
}

interface BarChart {
  bar_chart: CartesianCanvasChartSpec;
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
