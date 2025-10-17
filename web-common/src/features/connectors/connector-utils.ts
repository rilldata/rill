import type { V1AnalyzedConnector } from "../../runtime-client";

/**
 * Determines the correct icon key for a connector based on its configuration.
 * Special case: MotherDuck connectors use "motherduck" icon even though they have driver: duckdb
 */
export function getConnectorIconKeyForMotherDuck(
  connector: V1AnalyzedConnector,
): string {
  // Special case: MotherDuck connectors use md: path prefix
  if (connector.config?.path?.startsWith("md:")) {
    return "motherduck";
  }

  // Default: use the driver name
  return connector.driver?.name || "duckdb";
}
