export const OLAP_DRIVERS_WITHOUT_MODELING = ["clickhouse", "druid"];

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
      return `${database}.${databaseSchema}.${table}`;
    case "pinot":
      return `${table}`;
    default:
      throw new Error(`Unsupported OLAP connector: ${connector}`);
  }
}
