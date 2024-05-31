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
    // Remove the secrets
    .filter(([key]) => !secretKeys.includes(key))
    .map(([key, value]) => `${key}: "${value}"`)
    .join("\n");

  // Return the compiled YAML
  return `${topOfFile}\n\n` + compiledKeyValues;
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
