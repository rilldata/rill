import { V1DeploymentStatus } from "@rilldata/web-admin/client";

// Re-export shared utilities from web-common
export {
  formatConnectorName,
  formatEnvironmentName,
} from "@rilldata/web-common/features/resources/display-utils";

/**
 * Returns the Tailwind CSS class for a deployment status indicator dot.
 * Green for running, yellow for in-progress states, red for errors, gray for not deployed.
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
 * Returns a color name for a resource kind tag.
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
