import {
  V1DeploymentStatus,
  createAdminServiceGetProject,
} from "@rilldata/web-admin/client";

const PollTimeDuringReconcile = 1000;
const PollTimeDuringError = 5000;
const PollTimeWhenProjectReady = 60 * 1000;

export function useProjectDeploymentStatus(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1DeploymentStatus>(orgName, projName, {
    query: {
      select: (data) => {
        return data?.prodDeployment?.status;
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
