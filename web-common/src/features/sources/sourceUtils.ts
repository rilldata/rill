import { extractFileExtension } from "@rilldata/web-common/features/entity-management/file-path-utils";
import {
  ConnectorDriverPropertyType,
  type ConnectorDriverProperty,
  type V1ConnectorDriver,
  type V1Source,
} from "@rilldata/web-common/runtime-client";
import { makeDotEnvConnectorKey } from "../connectors/code-utils";
import { getDriverNameForConnector } from "../connectors/connectors-utils";
import { sanitizeEntityName } from "../entity-management/name-utils";
import { getConnectorSchema } from "./modal/connector-schemas";
import { findGroupedEnumKeys } from "../templates/schema-utils";

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
          getDriverNameForConnector(connector.name as string),
          key,
        )} }}"`;
      }

      if (key === "sql") {
        // For SQL, we want to use a multi-line string and add a dev section
        const sqlLines = value
          .split("\n")
          .map((line) => `  ${line}`)
          .join("\n");
        const devSqlLines = value
          .split("\n")
          .map((line) => `    ${line}`)
          .join("\n");
        return `${key}: |\n${sqlLines}\n\ndev:\n  ${key}: |\n${devSqlLines}\n      limit 10000`;
      }

      const isStringProperty = stringPropertyKeys.includes(key);
      if (isStringProperty) {
        return `${key}: "${value}"`;
      }

      return `${key}: ${value}`;
    })
    .join("\n");

  return (
    `${SOURCE_MODEL_FILE_TOP}\n\nconnector: ${getDriverNameForConnector(connector.name as string)}\n\n` +
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
 * Prepare connector form values before submission.
 * Handles special transformations like ClickHouse auth_method → managed.
 */
export function prepareConnectorFormData(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
): Record<string, unknown> {
  const processedValues = { ...formValues };

  // Get schema to check for grouped fields
  const schema = connector.name ? getConnectorSchema(connector.name) : null;

  if (schema) {
    // Find all grouped enum keys (auth_method, connection_method, etc.)
    const groupedEnumKeys = findGroupedEnumKeys(schema);

    if (groupedEnumKeys.length > 0) {
      // Collect all fields that should be included based on active selections
      const allowedFields = new Set<string>();

      // For each grouped enum, find which fields are in the active option's group
      for (const enumKey of groupedEnumKeys) {
        const enumValue = processedValues[enumKey] as string | undefined;
        const prop = schema.properties?.[enumKey];
        const groupedFields = prop?.["x-grouped-fields"];

        if (enumValue && groupedFields && groupedFields[enumValue]) {
          // Add all fields from the active group
          for (const fieldKey of groupedFields[enumValue]) {
            allowedFields.add(fieldKey);
          }
        }
      }

      // Also include fields that aren't controlled by any grouped enum (standalone fields)
      // Collect all fields that ARE controlled by some group
      const allGroupedFieldKeys = new Set<string>();
      for (const enumKey of groupedEnumKeys) {
        const prop = schema.properties?.[enumKey];
        const groupedFields = prop?.["x-grouped-fields"];
        if (groupedFields) {
          for (const fieldArray of Object.values(groupedFields)) {
            for (const fieldKey of fieldArray) {
              allGroupedFieldKeys.add(fieldKey);
            }
          }
        }
      }

      // Filter processedValues to only include allowed fields
      const filteredValues: Record<string, unknown> = {};
      for (const [key, value] of Object.entries(processedValues)) {
        // Include if:
        // - It's in the allowed fields for active groups, OR
        // - It's not controlled by any group (standalone field), OR
        // - It's a grouped enum key itself (we'll remove it later if needed)
        if (allowedFields.has(key) || !allGroupedFieldKeys.has(key)) {
          filteredValues[key] = value;
        }
      }

      // ClickHouse: translate auth_method to managed boolean BEFORE removing grouped enum keys
      if (connector.name === "clickhouse" && processedValues.auth_method) {
        const authMethod = processedValues.auth_method as string;

        if (authMethod === "rill-managed") {
          // Rill-managed: set managed=true, mode=readwrite
          filteredValues.managed = true;
          filteredValues.mode = "readwrite";
        } else if (authMethod === "self-managed") {
          // Self-managed: set managed=false
          filteredValues.managed = false;
        }
      }

      // ClickHouse Cloud: set managed=false, ssl will be in filteredValues if using parameters tab
      if (connector.name === "clickhousecloud") {
        filteredValues.managed = false;
        // Only set ssl=true if it's in the filtered values (i.e., using parameters tab)
        if ('ssl' in filteredValues) {
          filteredValues.ssl = true;
        }
      }

      // Replace with filtered values
      Object.keys(processedValues).forEach(key => delete processedValues[key]);
      Object.assign(processedValues, filteredValues);

      // Remove the grouped enum keys themselves - they're UI-only fields
      for (const enumKey of groupedEnumKeys) {
        delete processedValues[enumKey];
      }
    }
  } else {
    // No schema, handle ClickHouse auth_method the old way
    if (connector.name === "clickhouse" && processedValues.auth_method) {
      const authMethod = processedValues.auth_method as string;

      if (authMethod === "rill-managed") {
        // Rill-managed: set managed=true, mode=readwrite
        processedValues.managed = true;
        processedValues.mode = "readwrite";
      } else if (authMethod === "self-managed") {
        // Self-managed: set managed=false
        processedValues.managed = false;
      }

      // Remove the UI-only auth_method field
      delete processedValues.auth_method;
    }

    // ClickHouse Cloud: set managed=false and ssl=true (only in non-schema path)
    if (connector.name === "clickhousecloud") {
      processedValues.managed = false;
      processedValues.ssl = true;
    }
  }

  return processedValues;
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

  // Apply DuckDB rewrite logic FIRST (before stripping connector properties)
  // This is important for connectors like SQLite that need connector properties
  // to build the SQL query before they're removed.
  const [rewrittenConnector, rewrittenFormValues] = maybeRewriteToDuckDb(
    connector,
    processedValues,
  );

  // Strip connector configuration keys from the source form values to prevent
  // leaking connector-level fields (e.g., credentials) into the model file.
  if (connector.configProperties) {
    const connectorPropertyKeys = new Set(
      connector.configProperties.map((p) => p.key).filter(Boolean),
    );
    for (const key of Object.keys(rewrittenFormValues)) {
      if (connectorPropertyKeys.has(key)) {
        delete rewrittenFormValues[key];
      }
    }
  }

  // Also strip UI-only grouped enum keys (auth_method, connection_method, mode, etc.)
  const schema = connector.name ? getConnectorSchema(connector.name) : null;
  if (schema) {
    const groupedEnumKeys = findGroupedEnumKeys(schema);
    for (const key of groupedEnumKeys) {
      delete rewrittenFormValues[key];
    }
  }

  // Handle placeholder values for required source properties
  if (rewrittenConnector.sourceProperties) {
    for (const prop of rewrittenConnector.sourceProperties) {
      if (prop.key && prop.required && !(prop.key in rewrittenFormValues)) {
        if (prop.placeholder) {
          rewrittenFormValues[prop.key] = prop.placeholder;
        }
      }
    }
  }

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
