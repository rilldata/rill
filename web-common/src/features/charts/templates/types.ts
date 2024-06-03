interface TemplateSpecProperties {
  name: string;
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

interface KPITemplate {
  name: "kpi";
  measure: string;
  title: string;
}

interface TableTemplate {
  name: "table";
  columns: string[];
  rows: string[];
}

type ChartTemplates = LineChart | BarChart;

type TemplateRenderes = ChartTemplates | KPITemplate | TableTemplate;

export type TemplateSpec = TemplateSpecProperties & TemplateRenderes;
