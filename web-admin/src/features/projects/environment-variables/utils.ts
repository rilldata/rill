import { EnvironmentType, type VariableNames } from "./types";

/**
 *
 * - Environment it belongs to ("prod" or "dev").
 * - If empty, then the value is used as the fallback for all environments.
 */
export function getEnvironmentLabel(environment: string) {
  return environment === EnvironmentType.UNDEFINED ? "" : environment;
}

/**
 * Checks if a key would create a duplicate environment variable.
 *
 * Rules for environment variables:
 * 1. A key can only be tied to one environment configuration
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
  environment: string,
  key: string,
  variableNames: VariableNames,
): boolean {
  return variableNames.some((variable) => {
    // Only consider it a duplicate if the same key exists
    if (variable.name === key) {
      // If either the existing or new variable is for all environments, it's a duplicate
      if (
        variable.environment === EnvironmentType.UNDEFINED ||
        environment === EnvironmentType.UNDEFINED
      ) {
        return true;
      }

      // If trying to create in the same environment as existing variable
      if (variable.environment === environment) {
        return true;
      }
    }
    return false;
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
