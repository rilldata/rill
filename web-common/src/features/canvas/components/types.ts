import type { KPIGridSpec } from "@rilldata/web-common/features/canvas/components/kpi-grid";
import type { ChartConfig, ChartType } from "./charts/types";
import type { ImageSpec } from "./image";
import type { KPISpec } from "./kpi";
import type { MarkdownSpec } from "./markdown";
import type { PivotSpec, TableSpec } from "./pivot";
import type { ChartSpec } from "./charts";

// First, let's create a union type for all possible specs
export type ComponentSpec =
  | ChartSpec
  | PivotSpec
  | ImageSpec
  | TableSpec
  | KPISpec
  | KPIGridSpec
  | MarkdownSpec;

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
  | "table";

interface LineChart {
  line_chart: ChartConfig;
}

interface AreaChart {
  area_chart: ChartConfig;
}

interface BarChart {
  bar_chart: ChartConfig;
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
