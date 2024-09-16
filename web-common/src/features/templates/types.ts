interface DataProperties {
  metrics_view: string;
  filter?: string;
}

export interface ChartProperties {
  x: string;
  y: string;
  xLabel?: string;
  yLabel?: string;
  color?: string;
}

interface LineChart {
  line_chart: ChartProperties;
}

interface BarChart {
  bar_chart: ChartProperties;
}

interface StackedBarChart {
  stacked_bar_chart: ChartProperties;
}

export interface KPIProperties extends DataProperties {
  measure: string;
  time_range: string;
  comparison_range?: string;
}
export interface KPITemplateT {
  kpi: KPIProperties;
}

export interface TableProperties extends DataProperties {
  time_range: string;
  measures: string[];
  comparison_range?: string;
  row_dimensions?: string[];
  col_dimensions?: string[];
}
export interface TableTemplateT {
  table: TableProperties;
}

export interface MarkdownProperties {
  content: string;
  css?: { [key: string]: any };
}

export interface MarkdownTemplateT {
  markdown: MarkdownProperties;
}

export interface SelectProperties {
  valueField: string;
  labelField?: string;
  label?: string;
  tooltip?: string;
  placeholder?: string;
}

export interface SelectPropertiesT {
  select: SelectProperties;
}

export interface SwitchProperties {
  label: string;
  value: string;
  tooltip?: string;
}

export interface SwitchPropertiesT {
  switch: SwitchProperties;
}

export interface ImageProperties {
  url: string;
  css?: { [key: string]: any };
}

export interface ImageTemplateT {
  image: ImageProperties;
}

type ChartTemplates = LineChart | BarChart | StackedBarChart;

export type TemplateSpec =
  | ChartTemplates
  | KPITemplateT
  | TableTemplateT
  | MarkdownTemplateT
  | ImageTemplateT
  | SelectPropertiesT
  | SwitchPropertiesT;
