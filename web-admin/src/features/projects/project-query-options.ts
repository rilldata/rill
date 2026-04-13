import {
  V1DeploymentStatus,
  type RpcStatus,
  type V1GetProjectResponse,
} from "@rilldata/web-admin/client";
import { RUNTIME_ACCESS_TOKEN_DEFAULT_TTL } from "@rilldata/web-common/runtime-client/constants";
import type { CreateQueryOptions } from "@tanstack/svelte-query";

const PollTimeWhenProjectDeploymentPending = 1000;
const PollTimeWhenProjectDeploymentError = 5000;
const PollTimeWhenProjectDeploymentOk = RUNTIME_ACCESS_TOKEN_DEFAULT_TTL / 2; // Proactively refetch the JWT before it expires

export const baseGetProjectQueryOptions: Partial<
  CreateQueryOptions<V1GetProjectResponse, RpcStatus>
> = {
  gcTime: Math.min(RUNTIME_ACCESS_TOKEN_DEFAULT_TTL, 1000 * 60 * 5), // Make sure we don't keep a stale JWT in the cache
  refetchInterval: (query) => {
    const status = query.state.data?.deployment?.status;
    switch (status) {
      case V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING:
      case V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING:
        return PollTimeWhenProjectDeploymentPending;
      case V1DeploymentStatus.DEPLOYMENT_STATUS_ERRORED:
        return PollTimeWhenProjectDeploymentError;
      case V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING:
        return PollTimeWhenProjectDeploymentOk;
      default:
        return false;
    }
  },
  refetchIntervalInBackground: true, // Keep polling while the tab is hidden (e.g. deploy loader)
  refetchOnMount: true,
  refetchOnReconnect: true,
  refetchOnWindowFocus: true,
};
