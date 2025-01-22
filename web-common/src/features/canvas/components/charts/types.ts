import type { ComponentType, SvelteComponent } from "svelte";

export interface FieldConfig {
  field: string;
  label?: string;
  format?: string;
  showAxisTitle?: boolean; // Default is false
  type: "quantitative" | "ordinal" | "nominal" | "temporal" | "geojson";
  timeUnit?: string; // For temporal fields
}

export interface ChartConfig {
  metrics_view: string;
  x?: FieldConfig;
  y?: FieldConfig;
  color?: FieldConfig | string;
  tooltip?: FieldConfig;
  vl_config?: string;
}

export type ChartType = "line_chart" | "bar_chart" | "stacked_bar";

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
  type: "quantitative" | "temporal" | "nominal" | "ordinal";
}
