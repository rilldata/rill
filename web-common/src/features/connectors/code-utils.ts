import { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import {
  type V1ConnectorDriver,
  type ConnectorDriverProperty,
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceGetFile,
} from "../../runtime-client";
import { runtime } from "../../runtime-client/runtime-store";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import {
  getName,
  isNonStandardIdentifier,
} from "@rilldata/web-common/features/entity-management/name-utils";
import {
  getRuntimeServiceAnalyzeConnectorsQueryKey,
  getRuntimeServiceGetInstanceQueryKey,
  runtimeServiceAnalyzeConnectors,
  runtimeServiceGetInstance,
  runtimeServicePutFile,
} from "../../runtime-client";
import {
  getDriverNameForConnector,
  makeSufficientlyQualifiedTableName,
} from "./connectors-utils";

function yamlModelTemplate(driverName: string) {
  return `# Model YAML
# Reference documentation: https://docs.rilldata.com/developers/build/connectors/data-source/${driverName}

type: model
materialize: true

connector: {{ connector }}

sql: {{ sql }}{{ dev_section }}
`;
}

export function compileConnectorYAML(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
  options?: {
    fieldFilter?: (
      property:
        | ConnectorDriverProperty
        | { key?: string; type?: string; secret?: boolean; internal?: boolean },
    ) => boolean;
    orderedProperties?: Array<
      | ConnectorDriverProperty
      | { key?: string; type?: string; secret?: boolean }
    >;
    connectorInstanceName?: string;
    secretKeys?: string[];
    stringKeys?: string[];
    schema?: { properties?: Record<string, { "x-env-var-name"?: string }> };
  },
) {
  // Add instructions to the top of the file
  const driverName = getDriverNameForConnector(connector.name as string);
  const topOfFile = `# Connector YAML
# Reference documentation: https://docs.rilldata.com/developers/build/connectors/data-source/${driverName}

type: connector

driver: ${driverName}`;

  // Use the provided orderedProperties if available.
  let properties = options?.orderedProperties ?? [];

  // Optionally filter properties
  if (options?.fieldFilter) {
    properties = properties.filter(options.fieldFilter);
  }

  // Get the secret property keys
  const secretPropertyKeys = options?.secretKeys ?? [];

  // Get the string property keys
  const stringPropertyKeys = options?.stringKeys ?? [];

  // Compile key value pairs in the order of properties
  const compiledKeyValues = properties
    .filter((property) => {
      if (!property.key) return false;
      const value = formValues[property.key];
      if (value === undefined) return false;
      // Filter out empty strings for optional fields
      if (typeof value === "string" && value.trim() === "") return false;
      // For ClickHouse, exclude managed: false as it's the default behavior
      // When managed=false, it's the default self-managed mode and doesn't need to be explicit
      if (
        connector.name === "clickhouse" &&
        property.key === "managed" &&
        value === false
      )
        return false;
      return true;
    })
    .map((property) => {
      const key = property.key as string;
      const value = formValues[key] as string;

      const isSecretProperty = secretPropertyKeys.includes(key);
      if (isSecretProperty) {
        const placeholder = `{{ .env.${makeDotEnvConnectorKey(
          connector.name as string,
          key,
          undefined, // No conflict detection for YAML generation
          options?.schema,
        )} }}`;
        return `${key}: "${placeholder}"`;
      }

      const isStringProperty = stringPropertyKeys.includes(key);
      if (isStringProperty) {
        return `${key}: "${value}"`;
      }

      return `${key}: ${value}`;
    })
    .join("\n");

  // Return the compiled YAML
  return `${topOfFile}\n` + compiledKeyValues;
}

export async function updateDotEnvWithSecrets(
  queryClient: QueryClient,
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
  formType: "source" | "connector",
  connectorInstanceName?: string,
  opts?: {
    secretKeys?: string[];
    schema?: { properties?: Record<string, { "x-env-var-name"?: string }> };
  },
): Promise<string> {
  const instanceId = get(runtime).instanceId;

  // Invalidate the cache to ensure we get fresh .env content
  // This prevents overwriting credentials added by a previous step
  await queryClient.invalidateQueries({
    queryKey: getRuntimeServiceGetFileQueryKey(instanceId, { path: ".env" }),
  });

  // Get the existing .env file with fresh data
  let blob: string;
  let originalBlob: string;
  try {
    const file = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetFileQueryKey(instanceId, { path: ".env" }),
      queryFn: () => runtimeServiceGetFile(instanceId, { path: ".env" }),
    });
    blob = file.blob || "";
    originalBlob = blob; // Keep original for conflict detection
  } catch (error) {
    // Handle the case where the .env file does not exist
    if (error?.response?.data?.message?.includes("no such file")) {
      blob = "";
      originalBlob = "";
    } else {
      throw error;
    }
  }

  // Get the secret keys
  const secretKeys = opts?.secretKeys ?? [];

  // In reality, all connectors have secret keys, but this is a safeguard
  if (!secretKeys) {
    return blob;
  }

  // Update the blob with the new secrets
  // Use originalBlob for conflict detection so all secrets use consistent naming
  secretKeys.forEach((key) => {
    if (!key || !formValues[key]) {
      return;
    }

    const connectorSecretKey = makeDotEnvConnectorKey(
      connector.name as string,
      key,
      originalBlob,
      opts?.schema,
    );

    blob = replaceOrAddEnvVariable(
      blob,
      connectorSecretKey,
      formValues[key] as string,
    );
  });

  return blob;
}

export function replaceOrAddEnvVariable(
  existingEnvBlob: string,
  key: string,
  newValue: string,
): string {
  const lines = existingEnvBlob.split("\n");
  let keyFound = false;

  const updatedLines = lines.map((line) => {
    if (line.startsWith(`${key}=`)) {
      keyFound = true;
      return `${key}=${newValue}`;
    }
    return line;
  });

  if (!keyFound) {
    updatedLines.push(`${key}=${newValue}`);
  }

  const newBlob = updatedLines
    .filter((line, index) => !(line === "" && index === 0))
    .join("\n")
    .trim();

  return newBlob;
}

export function deleteEnvVariable(
  existingEnvBlob: string,
  key: string,
): string {
  const lines = existingEnvBlob.split("\n");
  const updatedLines = lines.filter((line) => !line.startsWith(`${key}=`));
  const newBlob = updatedLines
    .filter((line, index) => !(line === "" && index === 0))
    .join("\n")
    .trim();

  return newBlob;
}

/**
 * Get a generic ALL_CAPS environment variable name
 * If schema provides x-env-var-name, use it directly.
 * Otherwise: Generic properties (AWS, Google, etc.) use no prefix
 * Driver-specific properties use DriverName_PropertyKey format
 */
export function getGenericEnvVarName(
  driverName: string,
  propertyKey: string,
  schema?: { properties?: Record<string, { "x-env-var-name"?: string }> },
): string {
  // If schema provides explicit env var name, use it
  const field = schema?.properties?.[propertyKey];
  if (field?.["x-env-var-name"]) {
    return field["x-env-var-name"];
  }

  // Fallback: Generic properties that don't need a driver prefix
  const genericProperties = new Set([
    // Google Cloud credentials
    "google_application_credentials",
    // AWS credentials (used by S3, Athena, Redshift, etc.)
    "aws_access_key_id",
    "aws_secret_access_key",
    "aws_access_token",
    // Azure
    "azure_storage_connection_string",
    "azure_storage_key",
    "azure_storage_sas_token",
    "azure_storage_account",
  ]);

  // Convert property key to SCREAMING_SNAKE_CASE
  const propertyKeyUpper = propertyKey
    .replace(/([a-z])([A-Z])/g, "$1_$2")
    .replace(/[._-]+/g, "_")
    .toUpperCase();

  // If it's a generic property, return just the property name
  if (genericProperties.has(propertyKey.toLowerCase())) {
    return propertyKeyUpper;
  }

  // Otherwise, use DriverName_PropertyKey format
  const driverNameUpper = driverName
    .replace(/([a-z])([A-Z])/g, "$1_$2")
    .replace(/[._-]+/g, "_")
    .toUpperCase();

  return `${driverNameUpper}_${propertyKeyUpper}`;
}

/**
 * Check if an environment variable exists in the env blob
 */
export function envVarExists(envBlob: string, varName: string): boolean {
  const lines = envBlob.split("\n");
  return lines.some((line) => line.startsWith(`${varName}=`));
}

/**
 * Find the next available environment variable name by appending _1, _2, etc.
 */
export function findAvailableEnvVarName(
  envBlob: string,
  baseName: string,
): string {
  let varName = baseName;
  let counter = 1;

  while (envVarExists(envBlob, varName)) {
    varName = `${baseName}_${counter}`;
    counter++;
  }

  return varName;
}

export function makeDotEnvConnectorKey(
  driverName: string,
  key: string,
  existingEnvBlob?: string,
  schema?: { properties?: Record<string, { "x-env-var-name"?: string }> },
) {
  // Generate generic ALL_CAPS environment variable name
  const baseGenericName = getGenericEnvVarName(driverName, key, schema);

  // If no existing env blob is provided, just return the base generic name
  if (!existingEnvBlob) {
    return baseGenericName;
  }

  // Check for conflicts and append _# if necessary
  return findAvailableEnvVarName(existingEnvBlob, baseGenericName);
}

export async function updateRillYAMLWithOlapConnector(
  queryClient: QueryClient,
  newConnector: string,
): Promise<string> {
  // Get the existing rill.yaml file
  const instanceId = get(runtime).instanceId;
  const file = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceGetFileQueryKey(instanceId, {
      path: "rill.yaml",
    }),
    queryFn: () => runtimeServiceGetFile(instanceId, { path: "rill.yaml" }),
  });
  const blob = file.blob || "";

  // Update the blob with the new OLAP connector
  return replaceOlapConnectorInYAML(blob, newConnector);
}

/**
 * Update the `olap_connector` key in a YAML file.
 * This function uses a regex approach to preserve comments and formatting.
 */
export function replaceOlapConnectorInYAML(
  blob: string,
  newConnector: string,
): string {
  const olapConnectorRegex = /^olap_connector: .+$/m;

  if (olapConnectorRegex.test(blob)) {
    return blob.replace(olapConnectorRegex, `olap_connector: ${newConnector}`);
  } else {
    return `${blob}${blob !== "" ? "\n" : ""}olap_connector: ${newConnector}\n`;
  }
}

export async function createYamlModelFromTable(
  queryClient: QueryClient,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
): Promise<[string, string]> {
  const instanceId = get(runtime).instanceId;

  // Get driver name for makeSufficientlyQualifiedTableName
  const analyzeConnectorsQueryKey =
    getRuntimeServiceAnalyzeConnectorsQueryKey(instanceId);
  const analyzeConnectorsQueryFn = async () =>
    runtimeServiceAnalyzeConnectors(instanceId);
  const connectors = await queryClient.fetchQuery({
    queryKey: analyzeConnectorsQueryKey,
    queryFn: analyzeConnectorsQueryFn,
  });
  const analyzedConnector = connectors?.connectors?.find(
    (c) => c.name === connector,
  );
  if (!analyzedConnector) {
    throw new Error(`Could not find connector ${connector}`);
  }
  const driverName = analyzedConnector.driver?.name as string;

  // Get new model name
  const allNames = [
    ...fileArtifacts.getNamesForKind(ResourceKind.Source),
    ...fileArtifacts.getNamesForKind(ResourceKind.Model),
  ];
  const newModelName = getName(`${table}_model`, allNames);
  const newModelPath = `models/${newModelName}.yaml`;

  // Get sufficiently qualified table name
  const sufficientlyQualifiedTableName = makeSufficientlyQualifiedTableName(
    driverName,
    database,
    databaseSchema,
    table,
  );

  // Use the sufficiently qualified table name directly
  const selectStatement = `select * from ${sufficientlyQualifiedTableName}`;

  // NOTE: Redshift does not support LIMIT clauses in its UNLOAD data exports.
  const shouldIncludeDevSection = driverName !== "redshift";
  const devSection = shouldIncludeDevSection
    ? `\n\ndev:\n  sql: ${selectStatement} limit 10000`
    : "";

  const yamlContent = yamlModelTemplate(driverName)
    .replace("{{ connector }}", connector)
    .replace(/{{ sql }}/g, selectStatement)
    .replace("{{ dev_section }}", devSection);

  // Write the YAML file
  await runtimeServicePutFile(instanceId, {
    path: newModelPath,
    blob: yamlContent,
    createOnly: true,
  });

  // Invalidate relevant queries
  await queryClient.invalidateQueries({
    queryKey: ["runtimeServiceListFiles", instanceId],
  });

  return ["/" + newModelPath, newModelName];
}

export async function createSqlModelFromTable(
  queryClient: QueryClient,
  connector: string,
  database: string,
  databaseSchema: string,
  table: string,
  addDevLimit: boolean = true,
): Promise<[string, string]> {
  const instanceId = get(runtime).instanceId;

  // Get driver name
  const analyzeConnectorsQueryKey =
    getRuntimeServiceAnalyzeConnectorsQueryKey(instanceId);
  const analyzeConnectorsQueryFn = async () =>
    runtimeServiceAnalyzeConnectors(instanceId);
  const connectors = await queryClient.fetchQuery({
    queryKey: analyzeConnectorsQueryKey,
    queryFn: analyzeConnectorsQueryFn,
  });
  const analyzedConnector = connectors?.connectors?.find(
    (c) => c.name === connector,
  );
  if (!analyzedConnector) {
    throw new Error(`Could not find connector ${connector}`);
  }
  const driverName = analyzedConnector.driver?.name as string;

  // Determine whether the connector is the default OLAP connector
  const runtimeInstanceQueryKey =
    getRuntimeServiceGetInstanceQueryKey(instanceId);
  const runtimeInstanceQueryFn = async () =>
    runtimeServiceGetInstance(instanceId, { sensitive: true });
  const runtimeInstance = await queryClient.fetchQuery({
    queryKey: runtimeInstanceQueryKey,
    queryFn: runtimeInstanceQueryFn,
  });
  if (!runtimeInstance) {
    throw new Error(`Could not find runtime instance ${instanceId}`);
  }
  const isDefaultOLAPConnector =
    runtimeInstance?.instance?.olapConnector === connector;

  // Get new model name
  const allNames = [
    ...fileArtifacts.getNamesForKind(ResourceKind.Source),
    ...fileArtifacts.getNamesForKind(ResourceKind.Model),
  ];
  const newModelName = getName(`${table}_model`, allNames);
  const newModelPath = `models/${newModelName}.sql`;

  // Get sufficiently qualified table name
  const sufficientlyQualifiedTableName = makeSufficientlyQualifiedTableName(
    driverName,
    database,
    databaseSchema,
    table,
  );

  // Create model
  const topComments = `-- Model SQL\n-- Reference documentation: https://docs.rilldata.com/developers/build/connectors/data-source/${driverName}`;
  const connectorLine = `-- @connector: ${connector}`;
  const selectStatement = isNonStandardIdentifier(
    sufficientlyQualifiedTableName,
  )
    ? `select * from "${sufficientlyQualifiedTableName}"`
    : `select * from ${sufficientlyQualifiedTableName}`;
  const devLimit = "{{ if dev }} limit 100000 {{ end}}";

  let modelSQL = `${topComments}\n`;

  if (!isDefaultOLAPConnector) {
    modelSQL += `${connectorLine}\n`;
  }

  modelSQL += `\n${selectStatement}`;

  if (addDevLimit) {
    modelSQL += `\n${devLimit}`;
  }

  await runtimeServicePutFile(instanceId, {
    path: newModelPath,
    blob: modelSQL,
    createOnly: true,
  });

  eventBus.emit("notification", {
    message: `Queried ${table} in workspace`,
  });

  // Done
  return ["/" + newModelPath, newModelName];
}
