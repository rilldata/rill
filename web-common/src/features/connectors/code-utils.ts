import { V1ConnectorDriver } from "../../runtime-client";

export function compileConnectorYAML(
  connector: V1ConnectorDriver,
  formValues: Record<string, unknown>,
) {
  // Add instructions to the top of the file
  const topOfFile = `# Connector YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/connectors
  
type: connector

driver: ${connector.name}`;

  // Get the secret keys
  const secretKeys =
    connector.sourceProperties
      ?.filter((property) => property.secret)
      .map((property) => property.key) || [];

  // Compile key value pairs
  const compiledKeyValues = Object.entries(formValues)
    .filter(([key]) => !secretKeys.includes(key)) // Remove the secrets
    .map(([key, value]) => `${key}: "${value}"`)
    .join("\n");

  // Return the compiled YAML
  return `${topOfFile}\n` + compiledKeyValues;
}

export function updateDotEnvBlobWithNewSecrets(
  blob: string,
  connector: V1ConnectorDriver,
  formValues: Record<string, string>,
): string {
  const secretKeys = connector.sourceProperties
    ?.filter((property) => property.secret)
    .map((property) => property.key);

  if (!secretKeys) {
    return blob;
  }

  secretKeys.forEach((key) => {
    if (!key) {
      return;
    }

    blob = updateDotEnvBlobWithNewSecret(
      blob,
      makeDotEnvConnectorKey(connector.name as string, key),
      formValues[key],
    );
  });

  return blob;
}

export function updateDotEnvBlobWithNewSecret(
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

export function makeDotEnvConnectorKey(connectorName: string, key: string) {
  // Note: The connector name, not driver, is used. This enables configuring multiple connectors that use the same driver.
  return `connector.${connectorName}.${key}`;
}

/**
 * Update the `olap_connector` key in a Rill YAML file.
 * This function uses a regex approach to preserve comments and formatting.
 */
export function updateRillYAMLBlobWithNewOlapConnector(
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
