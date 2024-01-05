import {
  V1DeploymentStatus,
  createAdminServiceGetProject,
} from "@rilldata/web-admin/client";

export const PollTimeWhenProjectDeploymentPending = 1000;
export const PollTimeWhenProjectDeploymentError = 5000;
export const PollTimeWhenProjectDeployed = 60 * 1000;

export function useProjectDeploymentStatus(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1DeploymentStatus>(orgName, projName, {
    query: {
      select: (data) => {
        // There may not be a prodDeployment if the project is hibernating
        return (
          data?.prodDeployment?.status ||
          V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED
        );
      },
      refetchInterval: (data) => {
        switch (data) {
          case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
            return PollTimeWhenProjectDeploymentPending;

          case V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR:
          case V1DeploymentStatus.DEPLOYMENT_STATUS_UNSPECIFIED:
            return PollTimeWhenProjectDeploymentError;

          case V1DeploymentStatus.DEPLOYMENT_STATUS_OK:
            return PollTimeWhenProjectDeployed;

          default:
            return PollTimeWhenProjectDeployed;
        }
      },
    },
  });
}
