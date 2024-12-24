import type { Writable } from "svelte/store";
import type { ComponentInputParam } from "../inspector/types";
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

// The CanvasComponent interface is generic over the spec type.
export interface CanvasComponent<T> {
  /**
   * Svelte writable for the spec
   */
  specStore: Writable<T>;

  /**
   * Path in the YAML where the component is stored
   */
  pathInYAML: (string | number)[];

  /**
   * Minimum allowed size for the component
   * container on the canvas
   */
  minSize: ComponentSize;

  /**
   * The default size of the container when the component
   * is added to the canvas
   */
  defaultSize: ComponentSize;

  /**
   * The minimum condition needed for the spec to be valid
   * for the given component and to be rendered on the canvas
   */
  isValid(spec: T): boolean;

  /**
   * A map of input params which will be used in the visual
   * UI builder
   */
  inputParams(): Record<keyof T, ComponentInputParam>;

  /**
   * Get the spec when the component is added to the canvas
   */
  newComponentSpec(metrics_view: string, measure: string, dimension: string): T;

  /**
   * Update the spec store with the new values
   */
  updateProperty(key: string, value: unknown): Promise<void>;

  /**
   * Set the spec store with the new values
   */
  setSpec(spec: T): Promise<void>;
}

export interface ComponentCommonProperties {
  title?: string;
  description?: string;
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

interface BarChart {
  bar_chart: ChartConfig;
}

export type ChartTemplates = LineChart | BarChart;
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
