import {
  type V1Deployment,
  V1DeploymentStatus,
} from "@rilldata/web-admin/client";
import type { FilterGroup } from "@rilldata/web-common/components/table-toolbar";

export function getDeploymentStatusFilterGroup(
  statusFilter: string[],
  includeUnpublished = false,
) {
  return {
    label: "Status",
    key: "status",
    options: [
      { label: "Ready", value: "running" },
      { label: "Pending", value: "pending" },
      { label: "Error", value: "errored" },
      { label: "Stopped", value: "stopped" },
      ...(includeUnpublished
        ? [{ label: "Unpublished", value: "unpublished" }]
        : []),
    ],
    selected: statusFilter,
    defaultValue: [],
    multiSelect: true,
  } satisfies FilterGroup;
}

export function deploymentStatusFilterMatches(
  statusFilter: string[],
  d: V1Deployment | undefined,
): boolean {
  if (statusFilter.length === 0) return true;
  const s = d?.status ?? V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;
  return statusFilter.some((sel) => {
    switch (sel) {
      case "running":
        return s === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING;
      case "pending":
        return (
          s === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
          s === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING
        );
      case "errored":
        return s === V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED;
      case "stopped":
        return (
          s === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPED ||
          s === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING
        );
      case "unpublished":
        return s === V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;
      default:
        return false;
    }
  });
}
