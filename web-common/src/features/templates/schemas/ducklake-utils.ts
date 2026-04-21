import { ducklakeSchema } from "./ducklake";
import type { MultiStepFormSchema } from "./types";

/**
 * Compose a DuckDB `ATTACH` clause string (without the leading `ATTACH`
 * keyword) from the individual DuckLake parameter form fields.
 *
 * Example output:
 *   `'ducklake:duckdb_database.ducklake' AS my_ducklake (DATA_PATH 'files/', OVERRIDE_DATA_PATH true)`
 */
export function composeDuckLakeAttach(values: Record<string, unknown>): string {
  const identifier = composeCatalogIdentifier(values);
  if (!identifier) return "";

  const alias = stringValue(values.alias);

  const options: string[] = [];

  const dataPath = composeDataPath(values);
  if (dataPath) {
    options.push(`DATA_PATH '${escapeSqlString(dataPath)}'`);
  }

  const stringParams: Array<[string, string]> = [
    ["META_PARAMETER_NAME", "meta_parameter_name"],
    ["METADATA_CATALOG", "metadata_catalog"],
    ["METADATA_SCHEMA", "metadata_schema"],
    ["SNAPSHOT_TIME", "snapshot_time"],
    ["SNAPSHOT_VERSION", "snapshot_version"],
  ];
  for (const [sqlKey, formKey] of stringParams) {
    const v = stringValue(values[formKey]);
    if (v) options.push(`${sqlKey} '${escapeSqlString(v)}'`);
  }

  // METADATA_PARAMETERS is already an object literal like `{a: 1}`, so emit
  // it without wrapping quotes.
  const metadataParameters = stringValue(values.metadata_parameters);
  if (metadataParameters) {
    options.push(`METADATA_PARAMETERS ${metadataParameters}`);
  }

  const rowLimit = numberValue(values.data_inlining_row_limit);
  if (rowLimit !== undefined) {
    options.push(`DATA_INLINING_ROW_LIMIT ${rowLimit}`);
  }

  // Always emit boolean options so the user can see each configured advanced
  // setting reflected in the generated ATTACH clause.
  const boolParams: Array<[string, string]> = [
    ["OVERRIDE_DATA_PATH", "override_data_path"],
    ["CREATE_IF_NOT_EXISTS", "create_if_not_exists"],
    ["ENCRYPTED", "encrypted"],
    ["AUTOMATIC_MIGRATION", "automatic_migration"],
  ];
  for (const [sqlKey, formKey] of boolParams) {
    const v = values[formKey];
    if (typeof v === "boolean") {
      options.push(`${sqlKey} ${v ? "true" : "false"}`);
    }
  }

  const optionsStr = options.length ? ` (${options.join(", ")})` : "";
  const aliasStr = alias ? ` AS ${alias}` : "";
  return `'ducklake:${escapeSqlString(identifier)}'${aliasStr}${optionsStr}`;
}

/**
 * Build the portion of the DuckLake URI that follows `ducklake:`,
 * dispatching on the chosen `catalog_type`.
 */
function composeCatalogIdentifier(values: Record<string, unknown>): string {
  const type = stringValue(values.catalog_type) || "duckdb";

  switch (type) {
    case "duckdb":
      return stringValue(values.catalog_duckdb_path);

    case "sqlite": {
      const path = stringValue(values.catalog_sqlite_path);
      return path ? `sqlite:${path}` : "";
    }

    case "postgres": {
      const kv = keyValuePairs([
        ["dbname", values.catalog_postgres_dbname],
        ["host", values.catalog_postgres_host],
        ["port", values.catalog_postgres_port],
        ["user", values.catalog_postgres_user],
        ["password", values.catalog_postgres_password],
      ]);
      return kv ? `postgres:${kv}` : "";
    }

    case "mysql": {
      const kv = keyValuePairs([
        ["database", values.catalog_mysql_database],
        ["host", values.catalog_mysql_host],
        ["port", values.catalog_mysql_port],
        ["user", values.catalog_mysql_user],
        ["password", values.catalog_mysql_password],
      ]);
      return kv ? `mysql:${kv}` : "";
    }

    default:
      return "";
  }
}

/**
 * Return the storage path for the currently selected `data_path_type`,
 * or undefined when no path is set (so DATA_PATH is omitted from ATTACH).
 */
function composeDataPath(values: Record<string, unknown>): string | undefined {
  const type = stringValue(values.data_path_type) || "local";
  const key = `data_path_${type}`;
  const value = stringValue(values[key]);
  return value || undefined;
}

function keyValuePairs(entries: Array<[string, unknown]>): string {
  const parts: string[] = [];
  for (const [key, raw] of entries) {
    const v = stringValue(raw);
    if (v) parts.push(`${key}=${v}`);
  }
  return parts.join(" ");
}

/**
 * If the provided schema is the DuckLake schema and the user has selected the
 * "parameters" tab, synthesise an `attach` value from the parameter fields
 * and return a copy of `values` with that `attach` set. Otherwise returns the
 * original values unchanged.
 */
export function applyDuckLakeFormTransform(
  schema: MultiStepFormSchema | null | undefined,
  values: Record<string, unknown>,
): Record<string, unknown> {
  if (schema !== ducklakeSchema) return values;
  if (values.connection_mode !== "parameters") return values;
  const attach = composeDuckLakeAttach(values);
  return { ...values, attach };
}

/**
 * Re-inject the composed `attach` key into an already filtered value map.
 * The normal tab-group filter drops `attach` when the "parameters" tab is
 * active (since `attach` belongs to the "sql" tab group); callers use this
 * helper to restore it from the pre-filter source values.
 */
export function injectDuckLakeAttach(
  schema: MultiStepFormSchema | null | undefined,
  filteredValues: Record<string, unknown>,
  sourceValues: Record<string, unknown>,
): Record<string, unknown> {
  if (schema !== ducklakeSchema) return filteredValues;
  if (sourceValues.connection_mode !== "parameters") return filteredValues;
  const attach = sourceValues.attach;
  if (typeof attach !== "string" || !attach) return filteredValues;
  return { ...filteredValues, attach };
}

function stringValue(v: unknown): string {
  if (typeof v !== "string") return "";
  return v.trim();
}

function numberValue(v: unknown): number | undefined {
  if (typeof v === "number" && Number.isFinite(v)) return v;
  if (typeof v === "string" && v.trim() !== "") {
    const n = Number(v);
    if (Number.isFinite(n)) return n;
  }
  return undefined;
}

function escapeSqlString(v: string): string {
  return v.replace(/'/g, "''");
}
