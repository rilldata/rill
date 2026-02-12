import { extractFileExtension } from "@rilldata/web-common/features/entity-management/file-path-utils";
import type {
  V1ConnectorDriver,
  V1Source,
} from "@rilldata/web-common/runtime-client";
import { makeDotEnvConnectorKey } from "../connectors/code-utils";
import { sanitizeEntityName } from "../entity-management/name-utils";
import { getConnectorSchema } from "./modal/connector-schemas";
import {
  getSchemaFieldMetaList,
  getSchemaSecretKeys,
  getSchemaStringKeys,
} from "../templates/schema-utils";

// Helper text that we put at the top of every Model YAML file
function sourceModelFileTop(driverName: string) {
  return `# Model YAML
# Reference documentation: https://docs.rilldata.com/developers/build/connectors/data-source/${driverName}

type: model
materialize: true`;
}

export function compileSourceYAML(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
  opts?: {
    secretKeys?: string[];
    stringKeys?: string[];
    connectorInstanceName?: string;
    originalDriverName?: string;
  },
) {
  const schema = getConnectorSchema(connector.name ?? "");

  // Get the secret property keys
  const secretPropertyKeys =
    opts?.secretKeys ??
    (schema ? getSchemaSecretKeys(schema, { step: "source" }) : []);

  // Get the string property keys
  const stringPropertyKeys =
    opts?.stringKeys ??
    (schema ? getSchemaStringKeys(schema, { step: "source" }) : []);

  const formatSqlBlock = (sql: string, indent: string) =>
    `sql: |\n${sql
      .split("\n")
      .map((line) => `${indent}${line}`)
      .join("\n")}`;
  const trimSqlForDev = (sql: string) => sql.trim().replace(/;+\s*$/, "");

  // Compile key value pairs
  const compiledKeyValues = Object.keys(formValues)
    .filter((key) => {
      // For source files, exclude user-provided name since we use connector type
      if (key === "name") return false;
      const value = formValues[key];
      if (value === undefined) return false;
      // Filter out empty strings for optional fields
      if (typeof value === "string" && value.trim() === "") return false;
      return true;
    })
    .map((key) => {
      const value = formValues[key] as string;

      const isSecretProperty = secretPropertyKeys.includes(key);
      if (isSecretProperty) {
        // For source files, we include secret properties
        return `${key}: "{{ .env.${makeDotEnvConnectorKey(
          connector.name as string,
          key,
        )} }}"`;
      }

      if (key === "sql") {
        // For SQL, we want to use a multi-line string
        return formatSqlBlock(value, "  ");
      }

      const isStringProperty = stringPropertyKeys.includes(key);
      if (isStringProperty) {
        return `${key}: "${value}"`;
      }

      return `${key}: ${value}`;
    })
    .join("\n");

  const devSection =
    connector.implementsWarehouse &&
    connector.name !== "redshift" &&
    typeof formValues.sql === "string" &&
    formValues.sql.trim()
      ? `\n\ndev:\n  ${formatSqlBlock(
          `${trimSqlForDev(formValues.sql)} limit 10000`,
          "    ",
        )}`
      : "";

  // Use connector instance name if provided, otherwise fall back to driver name
  const connectorName = opts?.connectorInstanceName || connector.name;

  const driverName = opts?.originalDriverName || connector.name || "duckdb";
  return (
    `${sourceModelFileTop(driverName)}\n\nconnector: ${connectorName}\n\n` +
    compiledKeyValues +
    devSection
  );
}

export function compileLocalFileSourceYAML(path: string) {
  return `${sourceModelFileTop("local_file")}\n\nconnector: duckdb\nsql: "${buildDuckDbQuery(path)}"`;
}

function buildDuckDbQuery(path: string | undefined): string {
  const safePath = typeof path === "string" ? path : "";
  const extension = extractFileExtension(safePath);
  if (extensionContainsParts(extension, [".csv", ".tsv", ".txt"])) {
    return `select * from read_csv('${safePath}', auto_detect=true, ignore_errors=1, header=true)`;
  } else if (extensionContainsParts(extension, [".parquet"])) {
    return `select * from read_parquet('${safePath}')`;
  } else if (extensionContainsParts(extension, [".json", ".ndjson"])) {
    return `select * from read_json('${safePath}', auto_detect=true, format='auto')`;
  }

  return `select * from '${safePath}'`;
}

/**
 * Checks if a file extension '.v1.parquet.gz' contains parts like '.parquet'
 */
function extensionContainsParts(
  fileExtension: string,
  extensionParts: Array<string>,
) {
  for (const extension of extensionParts) {
    if (fileExtension.includes(extension)) return true;
  }
  return false;
}

export function inferSourceName(connector: V1ConnectorDriver, path: string) {
  if (
    !path ||
    path.endsWith("/") ||
    (connector.name === "gcs" && !path.startsWith("gs://")) ||
    (connector.name === "s3" && !path.startsWith("s3://")) ||
    (connector.name === "https" &&
      !path.startsWith("https://") &&
      !path.startsWith("http://"))
  )
    return;

  const slug = path
    .split("/")
    .filter((s: string) => s.length > 0)
    .pop();

  if (!slug) return;

  const fileName = slug.split(".").shift();

  if (!fileName) return;

  return sanitizeEntityName(fileName);
}

export function inferModelNameFromSQL(sql: string): string | undefined {
  if (!sql) return;
  const match = sql.match(/\bFROM\s+([^\s;,()]+)/i);
  if (!match) return;
  // Take the last segment if schema-qualified (e.g. schema.table)
  const raw = match[1]
    .replace(/[`"[\]]/g, "")
    .split(".")
    .pop();
  if (!raw) return;
  return sanitizeEntityName(raw);
}

export function getFileTypeFromPath(fileName) {
  if (!fileName.includes(".")) return "";
  const fileType = fileName.split(/[#?]/)[0].split(".").pop();

  if (!fileType) return "";

  if (fileType === "gz") {
    return fileName.split(".").slice(-2).shift();
  }

  return fileType;
}

/**
 * Detect the ClickHouse format name from a file path extension.
 * Maps common extensions to ClickHouse input format identifiers.
 */
function getClickHouseFormat(path: string): string {
  const extension = extractFileExtension(path);
  if (extensionContainsParts(extension, [".csv", ".tsv", ".txt"])) {
    return "CSVWithNames";
  } else if (extensionContainsParts(extension, [".parquet"])) {
    return "Parquet";
  } else if (extensionContainsParts(extension, [".json", ".ndjson"])) {
    return "JSONEachRow";
  }
  return "CSVWithNames";
}

/**
 * Build a multi-line ClickHouse SQL function call for readability.
 * Each argument is placed on its own indented line.
 */
function chFn(name: string, args: string[]): string {
  if (args.length <= 2) {
    return `SELECT * FROM ${name}(${args.map((a) => `'${a}'`).join(", ")})`;
  }
  const indentedArgs = args.map((a) => `  '${a}'`).join(",\n");
  return `SELECT * FROM ${name}(\n${indentedArgs}\n)`;
}

/**
 * Build a ClickHouse SQL query for reading data from various sources.
 * ClickHouse uses table functions like s3(), gcs(), url(), file() instead of DuckDB's read_* functions.
 * Credentials are referenced via {{ .env.connector.<name>.<key> }} templates that Rill resolves at runtime.
 */
function buildClickHouseQuery(
  connector: string,
  formValues: Record<string, unknown>,
  options?: { authMethod?: string },
): string {
  const isPublic = options?.authMethod === "public";
  const envRef = (key: string) =>
    `{{ .env.${makeDotEnvConnectorKey(connector, key)} }}`;

  switch (connector) {
    case "s3": {
      const path = typeof formValues.path === "string" ? formValues.path : "";
      const fmt = getClickHouseFormat(path);
      if (isPublic) {
        return chFn("s3", [path, fmt]);
      }
      return chFn("s3", [
        path,
        envRef("aws_access_key_id"),
        envRef("aws_secret_access_key"),
        fmt,
      ]);
    }
    case "gcs": {
      const path = typeof formValues.path === "string" ? formValues.path : "";
      const fmt = getClickHouseFormat(path);
      if (isPublic) {
        return chFn("gcs", [path, fmt]);
      }
      // ClickHouse GCS only supports HMAC keys (not JSON credentials)
      return chFn("gcs", [path, envRef("key_id"), envRef("secret"), fmt]);
    }
    case "azure": {
      const path = typeof formValues.path === "string" ? formValues.path : "";
      const fmt = getClickHouseFormat(path);
      const authMethod = options?.authMethod;
      if (isPublic) {
        return chFn("azureBlobStorage", [path, fmt]);
      }
      if (authMethod === "connection_string") {
        return chFn("azureBlobStorage", [
          envRef("azure_storage_connection_string"),
          path,
          fmt,
        ]);
      }
      if (authMethod === "sas_token") {
        return chFn("azureBlobStorage", [
          path,
          envRef("azure_storage_account"),
          envRef("azure_storage_sas_token"),
          fmt,
        ]);
      }
      // account_key (default)
      return chFn("azureBlobStorage", [
        path,
        envRef("azure_storage_account"),
        envRef("azure_storage_key"),
        fmt,
      ]);
    }
    case "https": {
      const path = typeof formValues.path === "string" ? formValues.path : "";
      return chFn("url", [path, getClickHouseFormat(path)]);
    }
    case "local_file": {
      const path = typeof formValues.path === "string" ? formValues.path : "";
      return chFn("file", [path, getClickHouseFormat(path)]);
    }
    case "sqlite": {
      const db = typeof formValues.db === "string" ? formValues.db : "";
      const table =
        typeof formValues.table === "string" ? formValues.table : "";
      return chFn("sqlite", [db, table]);
    }
    case "mysql": {
      const host = typeof formValues.host === "string" ? formValues.host : "";
      const port = typeof formValues.port === "string" ? formValues.port : "3306";
      const database =
        typeof formValues.database === "string" ? formValues.database : "";
      const user = typeof formValues.user === "string" ? formValues.user : "";
      const table =
        typeof formValues.table === "string" ? formValues.table : "";
      return chFn("mysql", [
        `${host}:${port}`,
        database,
        table,
        user,
        envRef("password"),
      ]);
    }
    case "postgres": {
      const host = typeof formValues.host === "string" ? formValues.host : "";
      const port = typeof formValues.port === "string" ? formValues.port : "5432";
      const dbname =
        typeof formValues.dbname === "string" ? formValues.dbname : "";
      const user = typeof formValues.user === "string" ? formValues.user : "";
      const table =
        typeof formValues.table === "string" ? formValues.table : "";
      return chFn("postgresql", [
        `${host}:${port}`,
        dbname,
        table,
        user,
        envRef("password"),
      ]);
    }
    default: {
      const path = typeof formValues.path === "string" ? formValues.path : "";
      return `SELECT * FROM '${path}'`;
    }
  }
}

/**
 * Compile a staging model YAML for ClickHouse.
 * Stages data from a warehouse (Snowflake/Redshift/BigQuery) through cloud storage
 * into ClickHouse.
 */
export function compileStagingYAML(
  formValues: Record<string, unknown>,
): string {
  const warehouse = String(formValues.warehouse ?? "snowflake");
  const sql = String(formValues.sql ?? "");
  const stagingConnector = String(formValues.staging_connector ?? "s3");
  const stagingPath = String(formValues.staging_path ?? "");
  const name = String(formValues.name ?? "");

  const formatSqlBlock = (s: string, indent: string) =>
    `sql: |\n${s
      .split("\n")
      .map((line) => `${indent}${line}`)
      .join("\n")}`;

  const lines: string[] = [
    `# Staging Model: ${warehouse} → ${stagingConnector} → ClickHouse`,
    `# This model extracts data from ${warehouse} using the provided SQL, stages it to ${stagingConnector}, and then reads it into ClickHouse.`,
    `For more details on staging with ClickHouse, see: https://docs.rilldata.com/developers/build/models/staging-models`,
    "",
    "type: model",
    "materialize: true",
    "",
    `connector: ${warehouse}`,
    formatSqlBlock(sql, "  "),
    "",
    "stage:",
    `  connector: ${stagingConnector}`,
    `  path: ${stagingPath}`,
    "",
    "output:",
    "  connector: clickhouse",
  ];

  return lines.join("\n");
}

/**
 * Convert applicable connectors to ClickHouse SQL when the OLAP engine is managed ClickHouse.
 * Parallel to maybeRewriteToDuckDb() but generates ClickHouse-dialect SQL.
 */
export function maybeRewriteToClickHouse(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
  options?: { connectorInstanceName?: string; authMethod?: string },
): [V1ConnectorDriver, Record<string, unknown>] {
  const connectorCopy = { ...connector };
  const originalConnectorName = connector.name ?? "";

  switch (originalConnectorName) {
    case "s3":
    case "gcs":
    case "azure":
    case "https":
    case "local_file":
      connectorCopy.name = "clickhouse";
      formValues.sql = buildClickHouseQuery(originalConnectorName, formValues, {
        authMethod: options?.authMethod,
      });
      delete formValues.path;
      break;
    case "sqlite":
      connectorCopy.name = "clickhouse";
      formValues.sql = buildClickHouseQuery(originalConnectorName, formValues, {
        authMethod: options?.authMethod,
      });
      delete formValues.db;
      delete formValues.table;
      break;
    case "mysql":
      connectorCopy.name = "clickhouse";
      formValues.sql = buildClickHouseQuery(originalConnectorName, formValues, {
        authMethod: options?.authMethod,
      });
      delete formValues.host;
      delete formValues.port;
      delete formValues.database;
      delete formValues.user;
      delete formValues.password;
      delete formValues["ssl-mode"];
      delete formValues.dsn;
      delete formValues.connection_mode;
      delete formValues.table;
      break;
    case "postgres":
      connectorCopy.name = "clickhouse";
      formValues.sql = buildClickHouseQuery(originalConnectorName, formValues, {
        authMethod: options?.authMethod,
      });
      delete formValues.host;
      delete formValues.port;
      delete formValues.dbname;
      delete formValues.user;
      delete formValues.password;
      delete formValues.sslmode;
      delete formValues.dsn;
      delete formValues.connection_mode;
      delete formValues.table;
      break;
  }

  // Clean up any remaining connector-step / UI-only fields that shouldn't
  // appear in the model YAML. The case handlers above delete the fields they
  // consume for SQL generation, but credentials and other connector fields
  // may still be present for source-only forms (where prepareSourceFormData
  // skips the connector-field stripping).
  if (
    connectorCopy.name === "clickhouse" &&
    originalConnectorName !== "clickhouse"
  ) {
    const schema = getConnectorSchema(originalConnectorName);
    if (schema?.properties) {
      for (const [key, prop] of Object.entries(schema.properties)) {
        if (key === "sql" || key === "name") continue;
        const step = prop["x-step"];
        // Remove connector-step fields, UI-only fields, and fields without x-step
        // (which default to connector behavior). Keep only source/explorer fields.
        if (step !== "source" && step !== "explorer") {
          delete formValues[key];
        }
      }
    }
  }

  return [connectorCopy, formValues];
}

/**
 * Dispatches to the appropriate rewrite function based on the OLAP engine.
 */
export function maybeRewriteForOlapEngine(
  olapEngine: string,
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
  options?: { connectorInstanceName?: string; authMethod?: string },
): [V1ConnectorDriver, Record<string, unknown>] {
  switch (olapEngine) {
    case "clickhouse":
      return maybeRewriteToClickHouse(connector, formValues, options);
    case "duckdb":
    default:
      return maybeRewriteToDuckDb(connector, formValues, options);
  }
}

/**
 * Convert applicable connectors to DuckDB. We do this to leverage DuckDB's native,
 * well-documented file reading capabilities.
 */
export function maybeRewriteToDuckDb(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
  options?: { connectorInstanceName?: string },
): [V1ConnectorDriver, Record<string, unknown>] {
  // Create a copy of the connector, so that we don't overwrite the original
  const connectorCopy = { ...connector };
  const connectorInstanceName =
    options?.connectorInstanceName?.trim() || undefined;
  const secretConnectorName = connectorInstanceName || connector.name || "";

  switch (connector.name) {
    case "s3":
    case "gcs":
    case "azure":
      // Ensure DuckDB creates a temporary secret for the original connector.
      if (secretConnectorName) {
        if (connectorInstanceName) {
          if (!formValues.create_secrets_from_connectors) {
            formValues.create_secrets_from_connectors = secretConnectorName;
          }
        } else {
          // When skipping connector creation, force the default driver name.
          formValues.create_secrets_from_connectors = secretConnectorName;
        }
      }
    // falls through to rewrite as DuckDB
    case "https":
      // HTTP sources are typically public; avoid surfacing secret wiring unless
      // the user is explicitly targeting a configured connector instance.
      if (connectorInstanceName && secretConnectorName) {
        if (!formValues.create_secrets_from_connectors) {
          formValues.create_secrets_from_connectors = secretConnectorName;
        }
      }
    // falls through to rewrite as DuckDB
    case "local_file":
      connectorCopy.name = "duckdb";

      formValues.sql = buildDuckDbQuery(formValues.path as string);
      delete formValues.path;

      break;
    case "sqlite":
      connectorCopy.name = "duckdb";

      formValues.sql = `SELECT * FROM sqlite_scan('${formValues.db as string}', '${
        formValues.table as string
      }');`;
      delete formValues.db;
      delete formValues.table;

      break;
  }

  return [connectorCopy, formValues];
}

/**
 * Process form data for sources, including OLAP-engine-aware rewrite logic and placeholder handling.
 * This serves as a single source of truth for both preview and submission.
 */
export function prepareSourceFormData(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
  options?: { connectorInstanceName?: string; olapEngine?: string },
): [V1ConnectorDriver, Record<string, unknown>] {
  // Create a copy of form values to avoid mutating the original
  const processedValues = { ...formValues };

  // Extract auth method before stripping (needed for ClickHouse SQL credential refs)
  const authMethod =
    typeof processedValues.auth_method === "string"
      ? processedValues.auth_method
      : undefined;

  // Never carry connector auth selection into the source/model layer.
  delete processedValues.auth_method;

  // Strip connector configuration keys from the source form values to prevent
  // leaking connector-level fields (e.g., credentials) into the model file.
  // For source-only forms (x-olap formType "source"), skip stripping here —
  // the rewrite function (e.g., maybeRewriteToClickHouse) needs these fields
  // for SQL generation and handles cleanup itself.
  const schema = getConnectorSchema(connector.name ?? "");
  const olapEngine = options?.olapEngine ?? "duckdb";
  const olapConfig = schema?.["x-olap"]?.[olapEngine];
  const isSourceOnlyForm = olapConfig?.formType === "source";
  const connectorPropertyKeys = new Set<string>();
  if (schema && !isSourceOnlyForm) {
    const connectorFields = getSchemaFieldMetaList(schema, {
      step: "connector",
    })
      .filter((field) => !field.internal)
      .map((field) => field.key);
    for (const key of connectorFields) {
      connectorPropertyKeys.add(key);
      delete processedValues[key];
    }
  }

  // Handle placeholder values for required source properties
  // Skip connector fields - they're handled by the connector, not the model
  if (schema) {
    const sourceFields = getSchemaFieldMetaList(schema, { step: "source" });
    for (const field of sourceFields) {
      // Don't fill placeholders for connector fields (even if they match source step)
      if (connectorPropertyKeys.has(field.key)) continue;
      if (field.required && !(field.key in processedValues)) {
        if (field.placeholder) {
          processedValues[field.key] = field.placeholder;
        }
      }
    }
  }

  // Apply OLAP-engine-aware rewrite logic (defaults to DuckDB)
  const [rewrittenConnector, rewrittenFormValues] = maybeRewriteForOlapEngine(
    olapEngine,
    connector,
    processedValues,
    { ...options, authMethod },
  );

  return [rewrittenConnector, rewrittenFormValues];
}

export function getFileExtension(source: V1Source): string {
  const path = String(source?.spec?.properties?.path).toLowerCase();
  if (path?.includes(".csv")) return "CSV";
  if (path?.includes(".parquet")) return "Parquet";
  if (path?.includes(".json")) return "JSON";
  if (path?.includes(".ndjson")) return "JSON";
  return "";
}

export function formatConnectorType(source: V1Source) {
  switch (source?.spec?.sourceConnector) {
    case "s3":
      return "S3";
    case "gcs":
      return "GCS";
    case "https":
      return "http(s)";
    case "local_file":
      return "Local file";
    default:
      return source?.state?.connector ?? "";
  }
}
