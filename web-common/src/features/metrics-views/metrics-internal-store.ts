import {
  CATEGORICALS,
  FLOATS,
} from "@rilldata/web-common/lib/duckdb-data-types";
import { DEFAULT_TIMEZONES } from "@rilldata/web-common/lib/time/config";
import type { V1StructType } from "@rilldata/web-common/runtime-client";
import { Document, parseDocument } from "yaml";
import { selectTimestampColumnFromSchema } from "./column-selectors";

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

export function initBlankDashboardYAML(dashboardTitle: string) {
  const metricsTemplate = `
# Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.

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

export function generateDashboardYAMLForTable(
  tableName: string,
  isModel: boolean,
  schema: V1StructType,
  dashboardTitle = "",
  defaultTimeRange = "",
) {
  const doc = new Document();

  doc.commentBefore = ` Visit https://docs.rilldata.com/reference/project-files to learn more about Rill project files.`;

  if (dashboardTitle) {
    doc.set("title", dashboardTitle);
  }

  if (isModel) {
    doc.set("model", tableName);
  } else {
    doc.set("table", tableName);
  }

  const timestampColumns = selectTimestampColumnFromSchema(schema);
  if (timestampColumns?.length) {
    doc.set("timeseries", timestampColumns[0]);
  } else {
    doc.set("timeseries", "");
  }

  if (defaultTimeRange) {
    doc.set("default_time_range", defaultTimeRange);
  }

  const fields = schema.fields;
  const metricsSeq = fields
    .filter((field) => {
      return FLOATS.has(field.type.code);
    })
    .map((field) => {
      return {
        label: "Sum(" + field.name + ")",
        expression: "sum(" + field.name + ")",
        name: "sum(" + field.name + ")",
        description: "Sum of " + capitalize(field.name),
        format_preset: "humanize",
        valid_percent_of_total: true,
      };
    });
  metricsSeq.unshift({
    label: "Total records",
    expression: "count(*)",
    name: "total_records",
    description: "Total number of records present",
    format_preset: "humanize",
    valid_percent_of_total: true,
  });
  doc.set("measures", metricsSeq);

  const dimensionSeq = fields
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

  const dimensionNode = doc.createNode(dimensionSeq);
  doc.set("dimensions", dimensionNode);

  doc.set("available_time_zones", DEFAULT_TIMEZONES);

  return doc.toString({ collectionStyle: "block" });
}
