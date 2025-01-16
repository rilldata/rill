import { EnvironmentType, type VariableNames } from "./types";

/**
 *
 * - Environment it belongs to ("prod" or "dev").
 * - If empty, then the value is used as the fallback for all environments.
 */
export function getEnvironmentType(environment: string) {
  return environment === EnvironmentType.UNDEFINED ? "" : environment;
}

/**
 * Checks if a key would create a duplicate environment variable.
 *
 * Rules for environment variables:
 * 1. A key can be tied to multiple environment configurations, e.g. my_key[dev] & my_key[prod]
 * 2. If a key exists in all environments (UNDEFINED), it cannot be created in dev or prod
 * 3. If a key exists in dev or prod, it cannot be created in all environments
 * 4. If a key exists in dev, it cannot be created in dev but CAN be created in prod
 * 5. Keys are case-insensitive across all environments
 */
export function isDuplicateKey(
  environment: string,
  key: string,
  existingVariables: VariableNames,
  currentKey?: string,
): boolean {
  const normalizedKey = key.toLowerCase();
  const normalizedCurrentKey = currentKey?.toLowerCase();

  return existingVariables.some((variable) => {
    const existingKey = variable.name.toLowerCase();

    // Skip if this is the current variable being edited
    if (
      normalizedCurrentKey &&
      existingKey === normalizedCurrentKey &&
      variable.environment === environment
    ) {
      return false;
    }

    // Safety: Compares the lowercase keys for case-insensitive comparison
    if (existingKey !== normalizedKey) {
      return false;
    }

    // If the existing variable is in UNDEFINED (all environments),
    // or we're trying to create in UNDEFINED, it's a duplicate
    if (
      variable.environment === EnvironmentType.UNDEFINED ||
      environment === EnvironmentType.UNDEFINED
    ) {
      return true;
    }

    // Otherwise, it's only a duplicate if it's in the same environment
    return variable.environment === environment;
  });
}

export function getCurrentEnvironment(
  isDevelopment: boolean,
  isProduction: boolean,
) {
  if (isDevelopment && isProduction) {
    return EnvironmentType.UNDEFINED;
  }

  if (isDevelopment) {
    return EnvironmentType.DEVELOPMENT;
  }

  if (isProduction) {
    return EnvironmentType.PRODUCTION;
  }

  return EnvironmentType.UNDEFINED;
}
