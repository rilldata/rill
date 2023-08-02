import { CATEGORICALS } from "@rilldata/web-common/lib/duckdb-data-types";
import type { V1Model } from "@rilldata/web-common/runtime-client";
import { parseDocument } from "yaml";
import { selectTimestampColumnFromSchema } from "./column-selectors";
import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";

export interface MetricsConfig extends MetricsParams {
  measures: MeasureEntity[];
  dimensions: DimensionEntity[];
}
export interface MetricsParams {
  display_name: string;
  title: string;
  timeseries: string;
  smallest_time_grain?: string;
  default_time_range?: string;
  model: string;
}
export interface MeasureEntity {
  name?: string;
  label?: string;
  expression?: string;
  description?: string;
  format_preset?: string;
  __GUID__?: string;
  __ERROR__?: string;
}
export interface DimensionEntity {
  name?: string;
  label?: string;
  property?: string;
  column?: string;
  description?: string;
  __ERROR__?: string;
}

const capitalize = (s) => s && s[0].toUpperCase() + s.slice(1);

export function initBlankDashboardYAML(dashboardName: string) {
  const metricsTemplate = `
# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

title: ""
model: ""
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
  - "Etc/UTC"
  - "America/Los_Angeles"
  - "America/New_York"
`;
  const template = parseDocument(metricsTemplate);
  template.set("title", dashboardName);
  return template.toString();
}

export function addQuickMetricsToDashboardYAML(yaml: string, model: V1Model) {
  const doc = parseDocument(yaml);
  doc.set("model", model.name);

  const timestampColumns = selectTimestampColumnFromSchema(model?.schema);
  if (timestampColumns?.length) {
    doc.set("timeseries", timestampColumns[0]);
  } else {
    doc.set("timeseries", "");
  }

  const measureNode = doc.createNode({
    label: "Total records",
    expression: "count(*)",
    name: "total_records",
    description: "Total number of records present",
    format_preset: "humanize",
    valid_percent_of_total: true,
  });
  doc.set("measures", [measureNode]);

  const fields = model.schema.fields;
  const diemensionSeq = fields
    .filter((field) => {
      return CATEGORICALS.has(field.type.code);
    })
    .map((field) => {
      return {
        name: field.name,
        label: capitalize(field.name),
        column: field.name,
        description: "",
      };
    });

  const dimensionNode = doc.createNode(diemensionSeq);
  doc.set("dimensions", dimensionNode);

  doc.set("available_time_zones", DEFAULT_TIMEZONES);

  return `# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

${doc.toString({ collectionStyle: "block" })}`;
}
