import type { V1AnalyzedConnector } from "../../runtime-client";

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
    case "mysql":
      // MySQL uses database.table format (no schema concept like PostgreSQL)
      if (database && database !== "default") {
        return `${database}.${table}`;
      }
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
    case "snowflake":
      return `/connector/snowflake/${connectorName}/${database}/${databaseSchema}/${table}`;
    case "bigquery":
      return `/connector/bigquery/${connectorName}/${database}/${databaseSchema}/${table}`;
    case "redshift":
      return `/connector/redshift/${connectorName}/${database}/${databaseSchema}/${table}`;
    case "athena":
      return `/connector/athena/${connectorName}/${database}/${databaseSchema}/${table}`;
    case "pinot":
      return `/connector/pinot/${connectorName}/${table}`;
    default:
      return null;
  }
}

/**
 * Determines the correct icon key for a connector based on its configuration.
 * Special cases:
 * - MotherDuck connectors use "motherduck" icon even though they have driver: duckdb
 * - ClickHouse Cloud connectors use "clickhousecloud" icon even though they have driver: clickhouse
 */
export function getConnectorIconKey(connector: V1AnalyzedConnector): string {
  // Special case: MotherDuck connectors use md: path prefix
  const path = connector.config?.path;
  if (typeof path === "string" && path.startsWith("md:")) {
    return "motherduck";
  }

  // Special case: ClickHouse Cloud connectors have "clickhouse.cloud" in host or dsn
  if (connector.driver?.name === "clickhouse") {
    const host = connector.config?.host;
    const dsn = connector.config?.dsn;

    if (
      (typeof host === "string" && host.includes("clickhouse.cloud")) ||
      (typeof dsn === "string" && dsn.includes("clickhouse.cloud"))
    ) {
      return "clickhousecloud";
    }
  }

  // Default: use the driver name
  return connector.driver?.name || "duckdb";
}

/**
 * Determines the driver name for a connector.
 * Special cases: MotherDuck uses "duckdb", Supabase uses "postgres".
 */
export function getDriverNameForConnector(connectorName: string): string {
  const driverMapping: Record<string, string> = {
    motherduck: "duckdb",
    supabase: "postgres",
  };
  return driverMapping[connectorName] ?? connectorName;
}
