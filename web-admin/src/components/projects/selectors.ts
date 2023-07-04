import {
  V1DeploymentStatus,
  createAdminServiceGetProject,
} from "@rilldata/web-admin/client";

export function getProjectPermissions(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, {
    query: {
      select: (data) => data?.projectPermissions,
    },
  });
}

const PollTimeDuringReconcile = 1000;
const PollTimeDuringError = 5000;
const PollTimeWhenProjectReady = 60 * 1000;

export function useProjectDeploymentStatus(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1DeploymentStatus>(orgName, projName, {
    query: {
      select: (data) => {
        // There may not be a prodDeployment if the project was hibernated
        return data?.prodDeployment?.status || V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED;
      },
      refetchInterval: (data) => {
        switch (data) {
          case "DEPLOYMENT_STATUS_PENDING":
          case "DEPLOYMENT_STATUS_RECONCILING":
            return PollTimeDuringReconcile;

          case "DEPLOYMENT_STATUS_ERROR":
          case "DEPLOYMENT_STATUS_UNSPECIFIED":
            return PollTimeDuringError;

          case "DEPLOYMENT_STATUS_OK":
            return PollTimeWhenProjectReady;

          default:
            return PollTimeWhenProjectReady;
        }
      },
    },
  });
}
