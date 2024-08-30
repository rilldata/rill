import type { QueryObserverOptions } from "@rilldata/svelte-query";
import {
  createAdminServiceGetProject,
  type RpcStatus,
  type V1Deployment,
  V1DeploymentStatus,
  type V1GetProjectResponse,
} from "@rilldata/web-admin/client";
import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants";
import { fixLocalhostRuntimePort } from "@rilldata/web-common/runtime-client/fix-localhost-runtime-port";

export function useProjectDeployment(orgName: string, projName: string) {
  return createAdminServiceGetProject<V1Deployment | undefined>(
    orgName,
    projName,
    undefined,
    {
      query: {
        select: (data) => {
          // There may not be a prodDeployment if the project is hibernating
          return data?.prodDeployment;
        },
      },
    },
  );
}

const PollTimeWhenProjectDeploymentPending = 1000;
const PollTimeWhenProjectDeploymentError = 5000;
const PollTimeWhenProjectDeploymentOk = RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2; // Proactively refetch the JWT before it expires

export const baseGetProjectQueryOptions: QueryObserverOptions<
  V1GetProjectResponse,
  RpcStatus
> = {
  cacheTime: Math.min(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL, 1000 * 60 * 5), // Make sure we don't keep a stale JWT in the cache
  refetchInterval: (data) => {
    switch (data?.prodDeployment?.status) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
        return PollTimeWhenProjectDeploymentPending;
      case V1DeploymentStatus.DEPLOYMENT_STATUS_ERROR:
        return PollTimeWhenProjectDeploymentError;
      case V1DeploymentStatus.DEPLOYMENT_STATUS_OK:
        return PollTimeWhenProjectDeploymentOk;
    }
  },
  refetchOnMount: true,
  refetchOnReconnect: true,
  refetchOnWindowFocus: true,
  select: (data: V1GetProjectResponse) => {
    if (data?.prodDeployment?.runtimeHost) {
      data.prodDeployment.runtimeHost = fixLocalhostRuntimePort(
        data.prodDeployment.runtimeHost,
      );
    }
    return data;
  },
};
