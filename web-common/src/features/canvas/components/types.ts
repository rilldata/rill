import type { ChartConfig, ChartType } from "./charts/types";
import type { ImageSpec } from "./image";
import type { KPISpec } from "./kpi";
import type { MarkdownSpec } from "./markdown";
import type { TableSpec } from "./table";

// First, let's create a union type for all possible specs
export type ComponentSpec =
  | ChartConfig
  | TableSpec
  | ImageSpec
  | KPISpec
  | MarkdownSpec;

export interface ComponentCommonProperties {
  title?: string;
  description?: string;
}

export interface ComponentFilterProperties {
  time_range?: string;
  comparison_range?: string;
  dimension_filters?: string;
}

export interface ComponentSize {
  width: number;
  height: number;
}

export type CanvasComponentType =
  | ChartType
  | "markdown"
  | "kpi"
  | "image"
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
export interface TableTemplateT {
  table: TableSpec;
}

export type TemplateSpec =
  | ChartTemplates
  | KPITemplateT
  | TableTemplateT
  | MarkdownTemplateT
  | ImageTemplateT;
