export type ChartType = "bar" | "stacked-bar" | "line";
export interface CanvasDashState {
  active: boolean;
}

export interface FieldConfig {
  field: string;
  type: "quantitative" | "ordinal" | "nominal" | "temporal" | "geojson";
  timeUnit?: string; // For temporal fields
}

export interface EncodingConfig {
  x?: FieldConfig;
  y?: FieldConfig;
  color?: FieldConfig;
}

export interface ChartTypeConfig {
  chartType: ChartType | null;
  data: EncodingConfig;
}
