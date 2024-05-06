export const OLAP_DRIVERS_WITHOUT_MODELING = ["clickhouse", "druid", "pinot"];

export function makeFullyQualifiedTableName(
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
) {
  switch (connector) {
    case "clickhouse":
      return `${databaseSchema}.${table}`;
    case "druid":
      return `${databaseSchema}.${table}`;
    case "duckdb":
      // return `${database}.${databaseSchema}.${table}`;
      // For now, only show the table name
      return table;
    case "pinot":
      return table;
    default:
      throw new Error(`Unsupported OLAP connector: ${connector}`);
  }
}

export function makeTablePreviewHref(
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
): string {
  switch (connector) {
    case "clickhouse":
      return `/connector/clickhouse/${databaseSchema}/${table}`;
    case "druid":
      return `/connector/druid/${databaseSchema}/${table}`;
    case "duckdb":
      return `/connector/duckdb/${database}/${databaseSchema}/${table}`;
    case "pinot":
      return `/connector/pinot/${table}`;
    default:
      throw new Error(`Unsupported connector: ${connector}`);
  }
}
