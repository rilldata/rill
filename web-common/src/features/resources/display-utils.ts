import type { V1Connector } from "@rilldata/web-common/runtime-client";

/**
 * Formats a connector name for display with proper capitalization.
 * Handles known connectors (duckdb, clickhouse, etc.) with correct casing.
 */
export function formatConnectorName(connector: string | undefined): string {
  if (!connector) return "\u2014";
  const lower = connector.toLowerCase();
  if (lower === "duckdb") return "DuckDB";
  if (lower === "ducklake") return "DuckLake";
  if (lower === "motherduck") return "MotherDuck";
  if (lower === "clickhouse") return "ClickHouse";
  if (lower === "mysql") return "MySQL";
  if (lower === "bigquery") return "BigQuery";
  if (lower === "openai") return "OpenAI";
  if (lower === "druid") return "Druid";
  if (lower === "pinot") return "Pinot";
  if (lower === "claude") return "Claude";
  if (lower === "gemini") return "Gemini";
  return connector.charAt(0).toUpperCase() + connector.slice(1);
}

/**
 * Returns the display label for the OLAP engine, including MotherDuck and
 * DuckLake detection and a management suffix (Rill-managed) where applicable.
 *
 * MotherDuck is detected by the connector's path starting with "md:" or a
 * token being configured. DuckLake is detected by `config.attach` referencing
 * a DuckLake catalog, or (as a fallback when configs are redacted or callers
 * only have the connector name) by a `ducklake` / `ducklake_*` name.
 *
 * Accepts a full V1Connector (with type + config) or a partial object holding
 * just a name — the latter is what web-local has access to via
 * `instance.olapConnector`.
 */
export function getOlapEngineLabel(connector: V1Connector | undefined): string {
  if (!connector) return "DuckDB";

  const name = connector.name ?? "";
  const lowerName = name.toLowerCase();
  const type = connector.type ?? "";

  const isDuckDB = type === "duckdb";
  const isMotherDuck =
    isDuckDB &&
    (String(connector.config?.path ?? "").startsWith("md:") ||
      !!connector.config?.token);
  const isDuckLake =
    !isMotherDuck &&
    ((isDuckDB &&
      String(connector.config?.attach ?? "").includes("ducklake:")) ||
      lowerName === "ducklake" ||
      lowerName.startsWith("ducklake_"));

  // When `type` is missing (e.g. web-local only has the olap connector name),
  // fall back to the name so "duckdb"/"clickhouse" still format correctly.
  const resolvedType = isMotherDuck
    ? "motherduck"
    : isDuckLake
      ? "ducklake"
      : type || name;
  const label = formatConnectorName(resolvedType);

  if (connector.provision) return `${label} (Rill-managed)`;
  return label;
}

/**
 * Formats an environment string for display with proper capitalization.
 * Handles common environment names (prod, dev, stage) and their variations.
 */
export function formatEnvironmentName(env: string | undefined): string {
  if (!env) return "Production";
  const lower = env.toLowerCase();
  if (lower === "prod" || lower === "production") return "Production";
  if (lower === "dev" || lower === "development") return "Development";
  if (lower === "stage" || lower === "staging") return "Staging";
  return env.charAt(0).toUpperCase() + env.slice(1);
}
