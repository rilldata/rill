import type { ComponentType, SvelteComponent } from "svelte";

export interface FieldConfig {
  field: string;
  label?: string;
  format?: string;
  type: "quantitative" | "ordinal" | "nominal" | "temporal" | "geojson";
  timeUnit?: string; // For temporal fields
}

export interface ChartConfig {
  metrics_view: string;
  x?: FieldConfig;
  y?: FieldConfig;
  color?: FieldConfig | string;
  tooltip?: FieldConfig;
  // Temp options
  time_range: string;
}

export type ChartType = "line_chart" | "bar_chart" | "stacked_bar";

export interface ChartMetadata {
  type: ChartType;
  icon: ComponentType<SvelteComponent>;
  title: string;
}
