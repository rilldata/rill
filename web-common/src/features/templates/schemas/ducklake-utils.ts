import {
  findAvailableEnvVarName,
  makeEnvVarKey,
  replaceOrAddEnvVariable,
} from "@rilldata/web-common/features/connectors/code-utils";
import { ducklakeSchema } from "./ducklake";
import type { MultiStepFormSchema } from "./types";

/**
 * Maps DuckLake password field keys (e.g. `catalog_postgres_password`) to the
 * resolved `.env` variable name used in the generated YAML. When present, the
 * composer emits `password={{ .env.<name> }}` instead of the raw value.
 */
export type DuckLakeSecretRefs = Record<string, string>;

export interface ComposeDuckLakeAttachOptions {
  secretRefs?: DuckLakeSecretRefs;
}

/**
 * Compose a DuckDB `ATTACH` clause string (without the leading `ATTACH`
 * keyword) from the individual DuckLake parameter form fields.
 *
 * Example output:
 *   `'ducklake:duckdb_database.ducklake' AS my_ducklake (DATA_PATH 'files/', OVERRIDE_DATA_PATH true)`
 */
export function composeDuckLakeAttach(
  values: Record<string, unknown>,
  opts?: ComposeDuckLakeAttachOptions,
): string {
  const identifier = composeCatalogIdentifier(values, opts?.secretRefs);
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
  // setting reflected in the generated ATTACH clause. `mode` is a boolean in
  // form state (true = read-only) and maps to DuckDB's `READ_ONLY` ATTACH flag.
  const boolParams: Array<[string, string]> = [
    ["READ_ONLY", "mode"],
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
function composeCatalogIdentifier(
  values: Record<string, unknown>,
  secretRefs?: DuckLakeSecretRefs,
): string {
  const type = stringValue(values.catalog_type) || "duckdb";

  switch (type) {
    case "duckdb":
      return stringValue(values.catalog_duckdb_path);

    case "sqlite": {
      const path = stringValue(values.catalog_sqlite_path);
      return path ? `sqlite:${path}` : "";
    }

    case "postgres": {
      const kv = keyValuePairs(
        [
          ["dbname", values.catalog_postgres_dbname],
          ["host", values.catalog_postgres_host],
          ["port", values.catalog_postgres_port],
          ["user", values.catalog_postgres_user],
          ["password", values.catalog_postgres_password],
        ],
        { password: fieldSecretRef(secretRefs, "catalog_postgres_password") },
      );
      return kv ? `postgres:${kv}` : "";
    }

    case "mysql": {
      const kv = keyValuePairs(
        [
          ["database", values.catalog_mysql_database],
          ["host", values.catalog_mysql_host],
          ["port", values.catalog_mysql_port],
          ["user", values.catalog_mysql_user],
          ["password", values.catalog_mysql_password],
        ],
        { password: fieldSecretRef(secretRefs, "catalog_mysql_password") },
      );
      return kv ? `mysql:${kv}` : "";
    }

    default:
      return "";
  }
}

function fieldSecretRef(
  secretRefs: DuckLakeSecretRefs | undefined,
  fieldKey: string,
): string | undefined {
  const envVarName = secretRefs?.[fieldKey];
  if (!envVarName) return undefined;
  return `{{ .env.${envVarName} }}`;
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

function keyValuePairs(
  entries: Array<[string, unknown]>,
  overrides: Record<string, string | undefined> = {},
): string {
  const parts: string[] = [];
  for (const [key, raw] of entries) {
    const override = overrides[key];
    if (override !== undefined) {
      // When an override is supplied (e.g. a templated secret reference), emit
      // it only if the user also provided a value for that field; otherwise
      // skip the pair so empty passwords do not inject a template ref.
      if (stringValue(raw)) parts.push(`${key}=${override}`);
      continue;
    }
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
 *
 * When `secretRefs` is supplied, password fields are emitted as
 * `{{ .env.<NAME> }}` template references rather than raw values, so the
 * generated ATTACH string stays free of plaintext secrets.
 */
export function applyDuckLakeFormTransform(
  schema: MultiStepFormSchema | null | undefined,
  values: Record<string, unknown>,
  opts?: ComposeDuckLakeAttachOptions,
): Record<string, unknown> {
  if (schema !== ducklakeSchema) return values;
  if (values.connection_mode === "parameters") {
    const attach = composeDuckLakeAttach(values, opts);
    return { ...values, attach };
  }
  // SQL mode: keep the user-typed value visible in the textarea, but strip
  // any `ATTACH ... ;` wrapper before it flows downstream into YAML.
  if (typeof values.attach === "string") {
    const cleaned = stripAttachKeyword(values.attach);
    if (cleaned !== values.attach) return { ...values, attach: cleaned };
  }
  return values;
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

/**
 * List of DuckLake password form-field keys whose values must be stored in
 * `.env` and referenced via template in the generated ATTACH clause.
 */
export const DUCKLAKE_SECRET_FIELD_KEYS = [
  "catalog_postgres_password",
  "catalog_mysql_password",
] as const;

/**
 * Resolve the `.env` variable names for DuckLake password fields, matching
 * the name `makeEnvVarKey` will use when writing secrets and compiling YAML.
 * Returns an empty object for non-DuckLake schemas so callers can pass the
 * result through unconditionally.
 */
export function buildDuckLakeSecretRefs(
  schema: MultiStepFormSchema | null | undefined,
  driverName: string,
  existingEnvBlob: string,
): DuckLakeSecretRefs {
  if (schema !== ducklakeSchema) return {};
  const refs: DuckLakeSecretRefs = {};
  for (const key of DUCKLAKE_SECRET_FIELD_KEYS) {
    refs[key] = makeEnvVarKey(driverName, key, existingEnvBlob, schema);
  }
  return refs;
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

/**
 * Catalog URI schemes whose bodies often carry raw credentials and should be
 * routed through `.env`. File-path catalogs (`sqlite:`, bare DuckDB files) are
 * intentionally omitted since they do not contain secrets.
 */
const DUCKLAKE_CATALOG_ENV_VAR_BASE: Record<string, string> = {
  postgres: "DUCKLAKE_POSTGRES",
  mysql: "DUCKLAKE_MYSQL",
  md: "DUCKLAKE_MOTHERDUCK",
};

const DUCKLAKE_CATALOG_URI_PATTERN = /'ducklake:(postgres|mysql|md):([^']*)'/g;

const ENV_TEMPLATE_ONLY_PATTERN = /^{{\s*\.env\.[^{}\s]+\s*}}$/;

export interface DuckLakeAttachExtraction {
  /** The attach string with raw catalog bodies replaced by `{{ .env.X }}` refs. */
  rewrittenAttach: string;
  /** Map of allocated env var name to the raw catalog body to persist in `.env`. */
  extractedSecrets: Record<string, string>;
}

/**
 * Extract credential-bearing catalog URIs from a raw DuckLake ATTACH string.
 *
 * For each `'ducklake:<driver>:<body>'` occurrence where driver is one of
 * `postgres`, `mysql`, or `md`, the entire body is extracted into a generic
 * env var (e.g. `DUCKLAKE_POSTGRES`) and replaced with a `{{ .env.<name> }}`
 * template reference. Conflicts against `existingEnvBlob` are resolved by
 * suffixing `_1`, `_2`, etc., matching `makeEnvVarKey`'s strategy.
 *
 * Idempotent: catalog bodies that already contain only a single env-template
 * reference are left unchanged so resubmitting a previously-extracted attach
 * does not re-wrap the value.
 */
export function extractDuckLakeAttachSecrets(
  attach: string,
  existingEnvBlob: string,
): DuckLakeAttachExtraction {
  if (!attach) return { rewrittenAttach: attach, extractedSecrets: {} };
  const extractedSecrets: Record<string, string> = {};
  // Track allocations across multiple matches in a single attach so two
  // postgres catalogs in one string don't collide on the same env var name.
  let reservedBlob = existingEnvBlob;

  const rewrittenAttach = attach.replace(
    DUCKLAKE_CATALOG_URI_PATTERN,
    (_match, driver: string, body: string) => {
      const trimmed = body.trim();
      if (!trimmed) return `'ducklake:${driver}:${body}'`;
      if (ENV_TEMPLATE_ONLY_PATTERN.test(trimmed)) {
        return `'ducklake:${driver}:${body}'`;
      }
      const base = DUCKLAKE_CATALOG_ENV_VAR_BASE[driver];
      const envVarName = findAvailableEnvVarName(reservedBlob, base);
      extractedSecrets[envVarName] = trimmed;
      reservedBlob = replaceOrAddEnvVariable(reservedBlob, envVarName, trimmed);
      return `'ducklake:${driver}:{{ .env.${envVarName} }}'`;
    },
  );
  return { rewrittenAttach, extractedSecrets };
}

/**
 * Return true when the schema + form values describe a DuckLake configuration
 * whose raw `attach` string should be scanned for catalog secrets. Parameters
 * mode is skipped since the composer already routes passwords through `.env`.
 */
export function shouldExtractDuckLakeAttachSecrets(
  schema: MultiStepFormSchema | null | undefined,
  values: Record<string, unknown>,
): boolean {
  if (schema !== ducklakeSchema) return false;
  return values.connection_mode !== "parameters";
}

export interface DuckLakePipelineResult {
  /**
   * Form values after Parameters-tab composition and SQL-tab catalog secret
   * extraction. Use these as the source of truth for YAML emission.
   */
  transformedValues: Record<string, unknown>;
  /**
   * Catalog secrets newly extracted from a raw `attach` string. Submit paths
   * append these to the `.env` blob; preview paths discard them.
   */
  extractedSecrets: Record<string, string>;
}

/**
 * Run the full DuckLake form-value pipeline: compose Parameters fields into
 * `attach`, route password fields through env-var refs, and extract catalog
 * secrets from a raw ATTACH string. No-op pass-through for non-DuckLake
 * schemas so callers can apply it unconditionally.
 *
 * Submit paths append `extractedSecrets` to their `.env` blob; preview paths
 * just use `transformedValues`.
 */
export function applyDuckLakeFormPipeline(
  schema: MultiStepFormSchema | null | undefined,
  formValues: Record<string, unknown>,
  opts: { connectorName: string; existingEnvBlob: string },
): DuckLakePipelineResult {
  if (schema !== ducklakeSchema) {
    return { transformedValues: formValues, extractedSecrets: {} };
  }

  const secretRefs = buildDuckLakeSecretRefs(
    schema,
    opts.connectorName,
    opts.existingEnvBlob,
  );
  let transformedValues = applyDuckLakeFormTransform(schema, formValues, {
    secretRefs,
  });

  let extractedSecrets: Record<string, string> = {};
  if (shouldExtractDuckLakeAttachSecrets(schema, transformedValues)) {
    const rawAttach = transformedValues.attach;
    if (typeof rawAttach === "string") {
      const result = extractDuckLakeAttachSecrets(
        rawAttach,
        opts.existingEnvBlob,
      );
      if (Object.keys(result.extractedSecrets).length > 0) {
        transformedValues = {
          ...transformedValues,
          attach: result.rewrittenAttach,
        };
        extractedSecrets = result.extractedSecrets;
      }
    }
  }

  return { transformedValues, extractedSecrets };
}

const DUCKLAKE_KNOWN_CATALOG_SCHEMES = new Set([
  "postgres",
  "mysql",
  "md",
  "sqlite",
]);

/**
 * Strip the optional `ATTACH [OR REPLACE] [IF NOT EXISTS]` prefix and a
 * trailing `;` from a user-pasted ATTACH statement. Rill emits the keyword
 * itself, so the stored form value should be just the clause body. The
 * textarea shows whatever the user typed; the wrapper is removed only when
 * the value flows into YAML, preview, or validation.
 */
export function stripAttachKeyword(value: string): string {
  return value
    .replace(/^\s*ATTACH\b\s*/i, "")
    .replace(/^(?:OR\s+REPLACE\s+)?(?:IF\s+NOT\s+EXISTS\s+)?/i, "")
    .replace(/;\s*$/, "")
    .trim();
}

/**
 * Structural validation for the raw DuckLake ATTACH SQL.
 *
 * Catches mistakes that we can identify without a full SQL parser:
 * leading `ATTACH` keyword, unbalanced quotes, missing `ducklake:` prefix
 * (when `TYPE DUCKLAKE` is also absent), empty catalog body, and unknown
 * catalog schemes. Returns one message per distinct issue; an empty array
 * means the string passes structural checks.
 */
export function validateDuckLakeAttach(attach: unknown): string[] {
  if (typeof attach !== "string") return [];
  // The user may paste a full `ATTACH ... ;` statement; strip the wrapper
  // before validating so we only check the clause body.
  const value = stripAttachKeyword(attach);
  if (!value) return [];

  const errors: string[] = [];

  const quoteCount = (value.match(/'/g) ?? []).length;
  if (quoteCount % 2 !== 0) {
    errors.push(
      "Unbalanced single quotes. Wrap the catalog URI in a single pair of quotes (e.g. 'ducklake:...').",
    );
  }

  // DuckLake accepts two equivalent forms: a `ducklake:` URI prefix, or a plain
  // URI paired with `(TYPE DUCKLAKE)` in the options. Only the first form has a
  // nested scheme we can validate (postgres:/mysql:/md:/sqlite:), so the scheme
  // checks below are skipped when `TYPE DUCKLAKE` is declared.
  const hasTypeDuckLakeOption = /\bTYPE\s+DUCKLAKE\b/i.test(value);

  const quotedBodies = [...value.matchAll(/'([^']*)'/g)].map((m) => m[1]);
  const duckLakeBodies = quotedBodies
    .filter((body) => body.trimStart().toLowerCase().startsWith("ducklake:"))
    .map((body) => body.trim());

  if (duckLakeBodies.length === 0) {
    if (!hasTypeDuckLakeOption) {
      errors.push(
        "Catalog URI must begin with `ducklake:`, or include `(TYPE DUCKLAKE)` in the options (e.g. 'ducklake:catalog.ducklake' or 'https://.../catalog.ducklake' AS x (TYPE DUCKLAKE)).",
      );
    }
    return errors;
  }

  for (const body of duckLakeBodies) {
    const afterDuckLake = body.slice("ducklake:".length);
    if (!afterDuckLake.trim()) {
      errors.push("Catalog URI has no value after `ducklake:`.");
      continue;
    }

    const schemeMatch = afterDuckLake.match(/^([a-zA-Z][a-zA-Z0-9_]*):/);
    if (!schemeMatch) continue;

    const scheme = schemeMatch[1].toLowerCase();
    const subBody = afterDuckLake.slice(schemeMatch[0].length).trim();

    if (!DUCKLAKE_KNOWN_CATALOG_SCHEMES.has(scheme)) {
      errors.push(
        `Unknown catalog scheme \`${scheme}:\`. Use one of: ${[...DUCKLAKE_KNOWN_CATALOG_SCHEMES].join(", ")}, or a file path for a DuckDB catalog.`,
      );
      continue;
    }

    if (!subBody) {
      errors.push(`\`${scheme}:\` catalog has no body.`);
    }
  }

  return errors;
}
