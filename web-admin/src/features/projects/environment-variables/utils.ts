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
 *
 * Examples:
 * - If test[dev] exists:
 *   - Cannot create test[dev] (same environment)
 *   - Cannot create test[all] (key already used)
 *   - CAN create test[prod] (different environment)
 * - If test[all] exists:
 *   - Cannot create test[dev] (key used in all environments)
 *   - Cannot create test[prod] (key used in all environments)
 *
 * @param environment - The target environment ("development", "production", or "undefined" for all)
 * @param key - The environment variable key to check
 * @param variableNames - Existing environment variables
 * @returns true if the key would create a duplicate, false otherwise
 */
export function isDuplicateKey(
  environment: EnvironmentType,
  key: string,
  variableNames: VariableNames,
  currentKey?: string,
): boolean {
  // If the key is the same as the current key, it's not a duplicate
  if (currentKey && key === currentKey) return false;

  const hasMatchingKey = (variable) => variable.name === key;
  const isInTargetEnvironment = (variable) =>
    environment === EnvironmentType.UNDEFINED ||
    variable.environment === environment ||
    variable.environment === EnvironmentType.UNDEFINED;

  // Check if the key already exists in the target environment or all environments
  return variableNames.some(
    (variable) => hasMatchingKey(variable) && isInTargetEnvironment(variable),
  );
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
