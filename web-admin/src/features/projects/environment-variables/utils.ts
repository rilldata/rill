import { EnvironmentType, type VariableNames } from "./types";

// FIXME: we should rename UNDEFINED to ALL for more clarity
// FOR NOW, we keep UNDEFINED based on the API
export function getEnvironmentLabel(environment: string) {
  return environment === EnvironmentType.UNDEFINED ? "all" : environment;
}

export function isDuplicateKey(
  environment: string,
  variableName: string,
  variableNames: VariableNames,
) {
  return variableNames.some((existingVariable) => {
    return (
      existingVariable.environment === environment &&
      existingVariable.name === variableName
    );
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
