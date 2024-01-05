import {
  V1DeploymentStatus,
  createAdminServiceGetProject,
} from "@rilldata/web-admin/client";

export const PollTimeDuringReconcile = 1000;
export const PollTimeDuringError = 5000;
export const PollTimeWhenProjectReady = 60 * 1000;

export function useProjectDeploymentStatus(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1DeploymentStatus>(orgName, projName, {
    query: {
      select: (data) => {
        // There may not be a prodDeployment if the project was hibernated
        return (
          data?.prodDeployment?.status ||
          V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED
        );
      },
      refetchInterval: (data) => {
        switch (data) {
          // case "DEPLOYMENT_STATUS_RECONCILING":
          case "DEPLOYMENT_STATUS_PENDING":
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
