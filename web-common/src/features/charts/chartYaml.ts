import type { V1ChartSpec } from "@rilldata/web-common/runtime-client";
import { Document, parse } from "yaml";

export function getChartYaml(
  vegaLite: string | undefined,
  resolver: string | undefined,
  resolverProperties: Record<string, any> | undefined,
) {
  const doc = new Document();
  doc.commentBefore = ` Chart YAML\n Reference documentation: https://docs.rilldata.com/reference/project-files/charts`;
  doc.set("kind", "chart");

  // TODO: more fields from resolverProperties
  if (resolver === "sql") {
    doc.set("data", { sql: (resolverProperties?.sql as string) ?? "" });
  } else if (resolver === "metrics_sql") {
    doc.set("data", { metrics_sql: (resolverProperties?.sql as string) ?? "" });
  } else if (resolver === "api") {
    doc.set("data", { api: (resolverProperties?.api as string) ?? "" });
  }

  doc.set("vega_lite", vegaLite ?? "{}");

  return doc.toString();
}

export function parseChartYaml(chartContent: string): V1ChartSpec {
  const chartYaml = parse(chartContent);

  let resolver = "";
  const resolverProperties: Record<string, string> = {};

  if (chartYaml.data?.sql) {
    resolver = "sql";
    resolverProperties.sql = chartYaml.data?.sql;
  } else if (chartYaml.data?.metrics_sql) {
    resolver = "metrics_sql";
    resolverProperties.sql = chartYaml.data?.metrics_sql;
  } else if (chartYaml.data?.api) {
    resolver = "api";
    resolverProperties.api = chartYaml.data?.api;
  }

  return {
    title: chartYaml.title,
    vegaLiteSpec: chartYaml.vegaLiteSpec,
    resolver,
    resolverProperties,
  };
}
