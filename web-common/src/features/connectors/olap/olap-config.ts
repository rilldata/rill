export const OLAP_DRIVERS_WITHOUT_MODELING = ["clickhouse", "druid", "pinot"];

export function makeTablePreviewHref(
  driver: string,
  connectorName: string,
  database: string,
  databaseSchema: string,
  table: string,
): string {
  switch (driver) {
    case "clickhouse":
      return `/connector/clickhouse/${connectorName}/${databaseSchema}/${table}`;
    case "druid":
      return `/connector/druid/${connectorName}/${databaseSchema}/${table}`;
    case "duckdb":
      return `/connector/duckdb/${connectorName}/${database}/${databaseSchema}/${table}`;
    case "pinot":
      return `/connector/pinot/${connectorName}/${table}`;
    default:
      throw new Error(`Unsupported connector: ${driver}`);
  }
}
