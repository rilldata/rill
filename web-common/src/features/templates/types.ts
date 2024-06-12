interface TemplateSpecProperties {
  name: string;
}

interface DataProperties {
  metric_view: string;
}

interface ChartProperties {
  x: string;
  y: string;
  xLabel?: string;
  yLabel?: string;
}

interface LineChart extends ChartProperties {
  name: "line";
}

interface BarChart extends ChartProperties {
  name: "bar";
}

export interface KPITemplateT extends DataProperties {
  name: "kpi";
  measure: string;
  time_range: string;
  comparison_range?: string;
}

export interface TableTemplateT extends DataProperties {
  name: "table";
  time_range: string;
  measures: string[];
  row_dimensions: string[];
  col_dimensions: string[];
}

type ChartTemplates = LineChart | BarChart;

type TemplateRenderes = ChartTemplates | KPITemplateT | TableTemplateT;

export type TemplateSpec = TemplateSpecProperties & TemplateRenderes;
