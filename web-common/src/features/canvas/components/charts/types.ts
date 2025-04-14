import type { ComponentType, SvelteComponent } from "svelte";
import type { ChartType } from "./";

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
