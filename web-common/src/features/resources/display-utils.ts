/**
 * Formats a connector name for display with proper capitalization.
 * Handles known connectors (duckdb, clickhouse, etc.) with correct casing.
 */
export function formatConnectorName(connector: string | undefined): string {
  if (!connector) return "\u2014";
  const lower = connector.toLowerCase();
  if (lower === "duckdb") return "DuckDB";
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
