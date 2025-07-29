import { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import {
  ConnectorDriverPropertyType,
  type V1ConnectorDriver,
  type ConnectorDriverProperty,
  getRuntimeServiceGetFileQueryKey,
  runtimeServiceGetFile,
} from "../../runtime-client";
import { runtime } from "../../runtime-client/runtime-store";

export function compileConnectorYAML(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
  options?: {
    fieldFilter?: (property: ConnectorDriverProperty) => boolean;
    orderedProperties?: ConnectorDriverProperty[];
  },
) {
  // Add instructions to the top of the file
  const topOfFile = `# Connector YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/connectors
  
type: connector

driver: ${connector.name}`;

  // Use the provided orderedProperties if available, otherwise fall back to configProperties/sourceProperties
  let properties =
    options?.orderedProperties ??
    connector.configProperties ??
    connector.sourceProperties ??
    [];

  // Optionally filter properties
  if (options?.fieldFilter) {
    properties = properties.filter(options.fieldFilter);
  }

  // Get the secret property keys
  const secretPropertyKeys =
    connector.configProperties
      ?.filter((property) => property.secret)
      .map((property) => property.key) || [];

  // Get the string property keys
  const stringPropertyKeys =
    connector.configProperties
      ?.filter(
        (property) => property.type === ConnectorDriverPropertyType.TYPE_STRING,
      )
      .map((property) => property.key) || [];

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
        return `${key}: "{{ .env.${makeDotEnvConnectorKey(
          connector.name as string,
          key,
        )} }}"`;
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
): Promise<string> {
  const instanceId = get(runtime).instanceId;

  // Get the existing .env file
  let blob: string;
  try {
    const file = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetFileQueryKey(instanceId, { path: ".env" }),
      queryFn: () => runtimeServiceGetFile(instanceId, { path: ".env" }),
    });
    blob = file.blob || "";
  } catch (error) {
    // Handle the case where the .env file does not exist
    if (error?.response?.data?.message?.includes("no such file")) {
      blob = "";
    } else {
      throw error;
    }
  }

  // Get the secret keys
  const properties =
    formType === "source"
      ? connector.sourceProperties
      : connector.configProperties;
  const secretKeys = properties
    ?.filter((property) => property.secret)
    .map((property) => property.key);

  // In reality, all connectors have secret keys, but this is a safeguard
  if (!secretKeys) {
    return blob;
  }

  // Update the blob with the new secrets
  secretKeys.forEach((key) => {
    if (!key || !formValues[key]) {
      return;
    }

    const connectorSecretKey = makeDotEnvConnectorKey(
      connector.name as string,
      key,
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

export function makeDotEnvConnectorKey(connectorName: string, key: string) {
  // Note: The connector name, not driver, is used. This enables configuring multiple connectors that use the same driver.
  return `connector.${connectorName}.${key}`;
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
