import type { MultiStepFormSchema } from "./types";

/**
 * Compose a DuckDB `ATTACH` clause string (without the leading `ATTACH`
 * keyword) from the individual DuckLake parameter form fields.
 *
 * Example output:
 *   `'ducklake:duckdb_database.ducklake' AS my_ducklake (DATA_PATH 'files/', OVERRIDE_DATA_PATH true)`
 */
export function composeDuckLakeAttach(values: Record<string, unknown>): string {
  const catalog = stringValue(values.catalog);
  if (!catalog) return "";

  const alias = stringValue(values.alias);

  const options: string[] = [];

  const stringParams: Array<[string, string]> = [
    ["DATA_PATH", "data_path"],
    ["META_PARAMETER_NAME", "meta_parameter_name"],
    ["METADATA_CATALOG", "metadata_catalog"],
    ["METADATA_PATH", "metadata_path"],
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
  if (rowLimit !== undefined && rowLimit > 0) {
    options.push(`DATA_INLINING_ROW_LIMIT ${rowLimit}`);
  }

  // Only emit boolean options when they differ from the DuckLake default —
  // otherwise they add clutter without changing behaviour.
  const boolParams: Array<[string, string, boolean]> = [
    ["OVERRIDE_DATA_PATH", "override_data_path", true],
    ["CREATE_IF_NOT_EXISTS", "create_if_not_exists", true],
    ["ENCRYPTED", "encrypted", false],
    ["AUTOMATIC_MIGRATION", "automatic_migration", false],
  ];
  for (const [sqlKey, formKey, defaultVal] of boolParams) {
    const v = values[formKey];
    if (typeof v === "boolean" && v !== defaultVal) {
      options.push(`${sqlKey} ${v ? "true" : "false"}`);
    }
  }

  const optionsStr = options.length ? ` (${options.join(", ")})` : "";
  const aliasStr = alias ? ` AS ${alias}` : "";
  return `'ducklake:${escapeSqlString(catalog)}'${aliasStr}${optionsStr}`;
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
  if (!schema || schema.title !== "DuckLake") return values;
  if (values.connection_mode !== "parameters") return values;
  const attach = composeDuckLakeAttach(values);
  return { ...values, attach };
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
