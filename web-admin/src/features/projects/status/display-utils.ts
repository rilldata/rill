import { V1DeploymentStatus } from "@rilldata/web-admin/client";

/**
 * Formats an environment string for display with proper capitalization.
 * Handles common environment names (prod, dev, stage) and their variations.
 * @param env - The environment name to format (e.g., "prod", "production")
 * @returns Formatted environment name (e.g., "Production")
 */
export function formatEnvironmentName(env: string | undefined): string {
  if (!env) return "Production";
  const lower = env.toLowerCase();
  if (lower === "prod" || lower === "production") return "Production";
  if (lower === "dev" || lower === "development") return "Development";
  if (lower === "stage" || lower === "staging") return "Staging";
  // Capitalize first letter for other environments
  return env.charAt(0).toUpperCase() + env.slice(1);
}

/**
 * Returns the Tailwind CSS class for a deployment status indicator dot.
 * Green for running, yellow for in-progress states, red for errors, gray for not deployed.
 * @param status - The deployment status
 * @returns Tailwind CSS class for the status dot background color
 */
export function getStatusDotClass(status: V1DeploymentStatus): string {
  switch (status) {
    case V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING:
      return "bg-green-500"; // Green - Ready
    case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
    case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
    case V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING:
    case V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING:
      return "bg-yellow-500"; // Yellow - In progress
    case V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED:
      return "bg-red-500"; // Red - Error
    default:
      return "bg-gray-400"; // Gray - Not deployed
  }
}

/**
 * Returns a human-readable label for a deployment status.
 * @param status - The deployment status
 * @returns Human-readable status label (e.g., "Ready", "Pending", "Error")
 */
export function getStatusLabel(status: V1DeploymentStatus): string {
  switch (status) {
    case V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING:
      return "Ready";
    case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
      return "Pending";
    case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
      return "Updating";
    case V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING:
      return "Stopping";
    case V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING:
      return "Deleting";
    case V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED:
      return "Error";
    case V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED:
      return "Stopped";
    case V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED:
      return "Deleted";
    default:
      return "Not deployed";
  }
}

/**
 * Formats a connector name for display with proper capitalization.
 * Handles known connectors (duckdb, clickhouse, etc.) with correct casing.
 * @param connector - The connector name to format
 * @returns Formatted connector name (e.g., "DuckDB", "ClickHouse") or em dash if undefined
 */
export function formatConnectorName(connector: string | undefined): string {
  if (!connector) return "—";
  // Capitalize first letter and clean up common names
  if (connector === "duckdb") return "DuckDB";
  if (connector === "clickhouse") return "ClickHouse";
  if (connector === "druid") return "Druid";
  if (connector === "pinot") return "Pinot";
  if (connector === "openai") return "OpenAI";
  if (connector === "claude") return "Claude";
  return connector.charAt(0).toUpperCase() + connector.slice(1);
}

/**
 * Returns a color name for a resource kind tag.
 * @param kind - The fully qualified resource kind (e.g., "rill.runtime.v1.Model")
 * @returns Color name for the tag (e.g., "blue", "green", "gray")
 */
export function getResourceKindTagColor(kind: string) {
  switch (kind) {
    case "rill.runtime.v1.MetricsView":
      return "blue";
    case "rill.runtime.v1.Model":
      return "green";
    case "rill.runtime.v1.Report":
      return "orange";
    case "rill.runtime.v1.Source":
      return "purple";
    case "rill.runtime.v1.Theme":
      return "magenta";
    default:
      return "gray";
  }
}
