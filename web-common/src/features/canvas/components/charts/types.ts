export interface FieldConfig {
  field: string;
  title?: string;
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
}
