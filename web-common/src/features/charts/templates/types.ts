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

interface KPITemplate extends DataProperties {
  name: "kpi";
  time_range: string;
  measure: string;
  title: string;
}

interface TableTemplate extends DataProperties {
  name: "table";
  time_range: string;
  measures: string[];
  row_dimensions: string[];
  col_dimensions: string[];
}

type ChartTemplates = LineChart | BarChart;

type TemplateRenderes = ChartTemplates | KPITemplate | TableTemplate;

export type TemplateSpec = TemplateSpecProperties & TemplateRenderes;
