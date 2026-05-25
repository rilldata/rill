/**
 * Get a generic ALL_CAPS environment variable name for a connector property.
 * If schema provides x-env-var-name, use it directly.
 * Otherwise uses DRIVER_NAME_PROPERTY_KEY format.
 *
 * @param driverName - The connector driver name (e.g., "clickhouse", "s3")
 * @param propertyKey - The property key (e.g., "password", "aws_access_key_id")
 * @param schema - Optional schema with x-env-var-name annotations
 * @returns The environment variable name in SCREAMING_SNAKE_CASE
 *
 * @example
 * getGenericEnvVarName("clickhouse", "password") // "CLICKHOUSE_PASSWORD"
 * getGenericEnvVarName("s3", "aws_access_key_id", s3Schema) // "AWS_ACCESS_KEY_ID" (from x-env-var-name)
 */
export function getGenericEnvVarName(
  driverName: string,
  propertyKey: string,
  schema: {
    properties?: Record<string, { "x-env-var-name"?: string }>;
  } | null = null,
): string {
  // If schema provides explicit env var name, use it
  const field = schema?.properties?.[propertyKey];
  if (field?.["x-env-var-name"]) {
    return field["x-env-var-name"];
  }

  // Convert property key to SCREAMING_SNAKE_CASE
  const propertyKeyUpper = propertyKey
    .replace(/([a-z])([A-Z])/g, "$1_$2")
    .replace(/[._-]+/g, "_")
    .toUpperCase();

  // Otherwise, use DriverName_PropertyKey format
  const driverNameUpper = driverName
    .replace(/([a-z])([A-Z])/g, "$1_$2")
    .replace(/[._-]+/g, "_")
    .toUpperCase();

  return `${driverNameUpper}_${propertyKeyUpper}`;
}
