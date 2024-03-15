import { Document } from "yaml";

export function getChartYaml(
  vegaLite: string | undefined,
  resolver: string | undefined,
  resolverProperties: Record<string, any> | undefined,
) {
  const doc = new Document();
  doc.set("kind", "chart");

  // TODO: more fields from resolverProperties
  if (resolver === "SQL") {
    doc.set("data", { sql: (resolverProperties?.sql as string) ?? "" });
  } else if (resolver === "MetricsSQL") {
    doc.set("data", { metrics_sql: (resolverProperties?.sql as string) ?? "" });
  } else if (resolver === "API") {
    doc.set("data", { api: (resolverProperties?.api as string) ?? "" });
  }

  doc.set(
    "vega_lite",
    JSON.stringify(JSON.parse(vegaLite ?? "{}"), null, 2).replace(/^/gm, "  "),
  );

  return doc.toString();
}
