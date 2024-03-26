import { Document } from "yaml";

export function getChartYaml(
  vegaLite: string | undefined,
  resolver: string | undefined,
  resolverProperties: Record<string, any> | undefined,
) {
  const doc = new Document();
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
