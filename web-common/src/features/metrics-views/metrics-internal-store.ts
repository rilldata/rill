import { parseDocument } from "yaml";

export function initBlankDashboardYAML(dashboardTitle: string) {
  const metricsTemplate = `
# Dashboard YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/dashboards

type: metrics_view

title: ""
table: ""
default_time_range: ""
smallest_time_grain: ""
timeseries: ""
measures:
  - label: "Total Records"
    expression: "count(*)"
dimensions:
  - name: dimension1
    label: First dimension
    column: dimension1
    description: ""
available_time_zones:
  - "UTC"
  - "America/Los_Angeles"
  - "America/New_York"
`;
  const template = parseDocument(metricsTemplate);
  template.set("title", dashboardTitle);
  return template.toString();
}
