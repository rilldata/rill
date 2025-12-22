import { extractFileExtension } from "@rilldata/web-common/features/entity-management/file-path-utils";
import {
  ConnectorDriverPropertyType,
  type ConnectorDriverProperty,
  type V1ConnectorDriver,
  type V1Source,
} from "@rilldata/web-common/runtime-client";
import { makeDotEnvConnectorKey } from "../connectors/code-utils";
import { sanitizeEntityName } from "../entity-management/name-utils";

// Helper text that we put at the top of every Model YAML file
const SOURCE_MODEL_FILE_TOP = `# Model YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/models

type: model
materialize: true`;

export function compileSourceYAML(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
) {
  // Get the secret property keys
  const secretPropertyKeys =
    connector.sourceProperties
      ?.filter((property) => property.secret)
      .map((property) => property.key) || [];

  // Get the string property keys
  const stringPropertyKeys =
    connector.sourceProperties
      ?.filter(
        (property) => property.type === ConnectorDriverPropertyType.TYPE_STRING,
      )
      .map((property) => property.key) || [];

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
        return `${key}: |\n  ${value
          .split("\n")
          .map((line) => `${line}`)
          .join("\n")}`;
      }

      const isStringProperty = stringPropertyKeys.includes(key);
      if (isStringProperty) {
        return `${key}: "${value}"`;
      }

      return `${key}: ${value}`;
    })
    .join("\n");

  return (
    `${SOURCE_MODEL_FILE_TOP}\n\nconnector: ${connector.name}\n\n` +
    compiledKeyValues
  );
}

export function compileLocalFileSourceYAML(path: string) {
  return `${SOURCE_MODEL_FILE_TOP}\n\nconnector: duckdb\nsql: "${buildDuckDbQuery(path)}"`;
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
): [V1ConnectorDriver, Record<string, unknown>] {
  // Create a copy of the connector, so that we don't overwrite the original
  const connectorCopy = { ...connector };

  switch (connector.name) {
    case "s3":
    case "gcs":
    case "https":
    case "azure":
    case "local_file":
      connectorCopy.name = "duckdb";

      formValues.sql = buildDuckDbQuery(formValues.path as string);
      delete formValues.path;

      connectorCopy.sourceProperties = [
        {
          key: "sql",
          type: ConnectorDriverPropertyType.TYPE_STRING,
        },
      ];

      break;
    case "sqlite":
      connectorCopy.name = "duckdb";

      formValues.sql = `SELECT * FROM sqlite_scan('${formValues.db as string}', '${
        formValues.table as string
      }');`;
      delete formValues.db;
      delete formValues.table;

      connectorCopy.sourceProperties = [
        {
          key: "sql",
          type: ConnectorDriverPropertyType.TYPE_STRING,
        },
      ];

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
): [V1ConnectorDriver, Record<string, unknown>] {
  // Create a copy of form values to avoid mutating the original
  const processedValues = { ...formValues };

  // Strip connector configuration keys from the source form values to prevent
  // leaking connector-level fields (e.g., credentials) into the model file.
  if (connector.configProperties) {
    const connectorPropertyKeys = new Set(
      connector.configProperties.map((p) => p.key).filter(Boolean),
    );
    for (const key of Object.keys(processedValues)) {
      if (connectorPropertyKeys.has(key)) {
        delete processedValues[key];
      }
    }
  }

  // Handle placeholder values for required source properties
  if (connector.sourceProperties) {
    for (const prop of connector.sourceProperties) {
      if (prop.key && prop.required && !(prop.key in processedValues)) {
        if (prop.placeholder) {
          processedValues[prop.key] = prop.placeholder;
        }
      }
    }
  }

  // Apply DuckDB rewrite logic
  const [rewrittenConnector, rewrittenFormValues] = maybeRewriteToDuckDb(
    connector,
    processedValues,
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

/**
 * Extracts initial form values from connector property specs, using the Default field if present.
 * @param properties Array of property specs (e.g., connector.configProperties)
 * @returns Object mapping property keys to their default values
 */
export function getInitialFormValuesFromProperties(
  properties: Array<ConnectorDriverProperty>,
) {
  const initialValues: Record<string, any> = {};
  for (const prop of properties) {
    // Only set if default is not undefined/null/empty string
    if (
      prop.key &&
      prop.default !== undefined &&
      prop.default !== null &&
      prop.default !== ""
    ) {
      let value: any = prop.default;
      if (prop.type === ConnectorDriverPropertyType.TYPE_NUMBER) {
        // NOTE: store number type prop as String, not Number, so that we can use the same form for both number and string properties
        // See `yupSchemas.ts` for more details
        value = String(value);
      } else if (prop.type === ConnectorDriverPropertyType.TYPE_BOOLEAN) {
        value = value === "true" || value === true;
      }
      initialValues[prop.key] = value;
    }
  }
  return initialValues;
}
