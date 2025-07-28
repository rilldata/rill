export const OLAP_DRIVERS_WITHOUT_MODELING = ["clickhouse", "druid", "pinot"];

export function makeFullyQualifiedTableName(
  driver: string,
  database: string,
  databaseSchema: string,
  table: string,
) {
  switch (driver) {
    case "clickhouse":
      return `${databaseSchema}.${table}`;
    case "druid":
      return `${databaseSchema}.${table}`;
    case "duckdb":
      return `${database}.${databaseSchema}.${table}`;
    case "pinot":
      return table;
    // Non-OLAP connectors: use standard database.schema.table format
    default:
      if (database && databaseSchema) {
        return `${database}.${databaseSchema}.${table}`;
      } else if (databaseSchema) {
        return `${databaseSchema}.${table}`;
      } else {
        return table;
      }
  }
}

/**
 * Returns a sufficiently qualified table name for the given OLAP connector.
 * Notably, the table name is *not* qualified with the database and databaseSchema when they are the default values for the given connector.
 */
export function makeSufficientlyQualifiedTableName(
  driver: string,
  database: string,
  databaseSchema: string,
  table: string,
) {
  switch (driver) {
    case "clickhouse":
      if (databaseSchema === "default" || databaseSchema === "") return table;
      return `${databaseSchema}.${table}`;
    case "druid":
      // TODO
      return table;
    case "duckdb":
      if (database === "main_db" || database === "") {
        if (databaseSchema === "main" || databaseSchema === "") return table;
        return `${databaseSchema}.${table}`;
      }
      return `${database}.${databaseSchema}.${table}`;
    case "pinot":
      // TODO
      return table;
    // Non-OLAP connectors: use standard qualification logic
    default:
      if (
        database &&
        databaseSchema &&
        database !== "default" &&
        databaseSchema !== "default"
      ) {
        return `${database}.${databaseSchema}.${table}`;
      } else if (databaseSchema && databaseSchema !== "default") {
        return `${databaseSchema}.${table}`;
      } else {
        return table;
      }
  }
}

export function makeTablePreviewHref(
  driver: string,
  connectorName: string,
  database: string,
  databaseSchema: string,
  table: string,
): string | null {
  switch (driver) {
    case "clickhouse":
      return `/connector/clickhouse/${connectorName}/${databaseSchema}/${table}`;
    case "druid":
      return `/connector/druid/${connectorName}/${databaseSchema}/${table}`;
    case "duckdb":
      return `/connector/duckdb/${connectorName}/${database}/${databaseSchema}/${table}`;
    case "pinot":
      return `/connector/pinot/${connectorName}/${table}`;
    // Non-OLAP connectors: table preview not implemented yet
    default:
      return null;
  }
}
