import { createAdminServiceGetProject } from "@rilldata/web-admin/client";

const PollTimeDuringReconcile = 1000;
const PollTimeDuringError = 5000;
const PollTimeWhenProjectReady = 60 * 1000;

export function useProject(orgName: string, projName: string) {
  return createAdminServiceGetProject(orgName, projName, {
    query: {
      refetchInterval: (data) => {
        console.log(data?.prodDeployment?.status);
        switch (data?.prodDeployment?.status) {
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
