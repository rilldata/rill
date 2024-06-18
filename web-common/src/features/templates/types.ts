interface DataProperties {
  metric_view: string;
}

interface ChartProperties {
  x: string;
  y: string;
  xLabel?: string;
  yLabel?: string;
}

interface LineChart {
  line_chart: ChartProperties;
}

interface BarChart extends ChartProperties {
  bar_chart: ChartProperties;
}

interface KPIProperties extends DataProperties {
  measure: string;
  time_range: string;
  comparison_range?: string;
}
export interface KPITemplateT {
  kpi: KPIProperties;
}

interface TableProperties extends DataProperties {
  time_range: string;
  measures: string[];
  row_dimensions: string[];
  col_dimensions: string[];
}
export interface TableTemplateT {
  table: TableProperties;
}

type ChartTemplates = LineChart | BarChart;

export type TemplateSpec = ChartTemplates | KPITemplateT | TableTemplateT;
