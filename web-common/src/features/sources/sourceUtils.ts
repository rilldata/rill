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
    existingEnvBlob?: string;
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
        return `${key}: '{{ env "${makeDotEnvConnectorKey(
          connector.name as string,
          key,
          opts?.existingEnvBlob,
          schema ?? undefined,
        )}" }}'`;
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
 * Process form data for sources, including DuckDB rewrite logic and placeholder handling.
 * This serves as a single source of truth for both preview and submission.
 */
export function prepareSourceFormData(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
  options?: { connectorInstanceName?: string },
): [V1ConnectorDriver, Record<string, unknown>] {
  // Create a copy of form values to avoid mutating the original
  const processedValues = { ...formValues };

  // Never carry connector auth selection into the source/model layer.
  delete processedValues.auth_method;

  // Strip connector configuration keys from the source form values to prevent
  // leaking connector-level fields (e.g., credentials) into the model file.
  const schema = getConnectorSchema(connector.name ?? "");
  const connectorPropertyKeys = new Set<string>();
  if (schema) {
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

  // Apply DuckDB rewrite logic
  const [rewrittenConnector, rewrittenFormValues] = maybeRewriteToDuckDb(
    connector,
    processedValues,
    options,
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
