import {
  V1DeploymentStatus,
  type V1ProjectHealth,
} from "@rilldata/web-admin/client";

export function isProjectHealthy(p: V1ProjectHealth): boolean {
  return (
    p.deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING &&
    (p.parseErrorCount ?? 0) === 0 &&
    (p.reconcileErrorCount ?? 0) === 0
  );
}

export function hasProjectErrors(p: V1ProjectHealth): boolean {
  return (
    p.deploymentStatus === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED ||
    (p.parseErrorCount ?? 0) > 0 ||
    (p.reconcileErrorCount ?? 0) > 0
  );
}
