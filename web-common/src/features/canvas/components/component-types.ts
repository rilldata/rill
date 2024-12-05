// TODO: Make it more human friendly and readable, along
// with perc, relative sizes etc.
export interface PositionDef {
  x: number;
  y: number;
  width: number;
  height: number;
}

interface ComponentCommonProperties {
  position: PositionDef;
  metricViewName: string;
  title?: string;
  subtitle?: string;
}

export interface MarkdownComponent extends ComponentCommonProperties {
  content: string;
}

export interface KPIComponent extends ComponentCommonProperties {
  measure: string;
  timeRange: string;
  showSparkline: boolean;
  comparisonRange?: string;
}

export type CanvasComponent = MarkdownComponent | KPIComponent;
