import { V1DeploymentStatus } from "../../client";

export function getDeploymentStatusText(status: V1DeploymentStatus) {
  switch (status) {
    case V1DeploymentStatus.DEPLOYMENT_STATUS_OK:
      return "Live";
    case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
      return "Pending";
    case V1DeploymentStatus.DEPLOYMENT_STATUS_RECONCILING:
      return "Reconciling";
    case V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR:
      return "Error";
    default:
      return "Unknown";
  }
}
