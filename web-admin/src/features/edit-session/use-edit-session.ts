import {
  createAdminServiceCreateDeployment,
  createAdminServiceListDeployments,
  getAdminServiceListDeploymentsQueryKey,
  V1DeploymentStatus,
  type AdminServiceListDeploymentsParams,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { derived, type Readable } from "svelte/store";

const DEV_DEPLOYMENT_POLL_INTERVAL = 2000;

const DEV_DEPLOYMENTS_PARAMS: AdminServiceListDeploymentsParams = {
  environment: "dev",
};

/**
 * Lists dev deployments for a project, polling while any are in a transitional state.
 */
export function useDevDeployments(org: string, project: string) {
  return createAdminServiceListDeployments(
    org,
    project,
    DEV_DEPLOYMENTS_PARAMS,
    {
      query: {
        refetchInterval: (query) => {
          const deployments = query.state.data?.deployments;
          if (!deployments?.length) return false;
          const hasTransitional = deployments.some((d) =>
            isTransitionalStatus(d.status),
          );
          return hasTransitional ? DEV_DEPLOYMENT_POLL_INTERVAL : false;
        },
      },
    },
  );
}

/**
 * Returns the first active dev deployment for the project (if any).
 * "Active" means not stopped, deleted, or errored.
 */
export function useActiveDevDeployment(
  org: string,
  project: string,
): Readable<{ data: V1Deployment | null; isLoading: boolean }> {
  const deploymentsQuery = useDevDeployments(org, project);

  return derived(deploymentsQuery, ($query) => {
    if ($query.isLoading) {
      return { data: null, isLoading: true };
    }
    const active =
      $query.data?.deployments?.find((d) => isActiveDeployment(d)) ?? null;
    return { data: active, isLoading: false };
  });
}

/**
 * Mutation to create a dev deployment with editable=true.
 */
export function useCreateDevDeployment() {
  return createAdminServiceCreateDeployment({
    mutation: {
      onSuccess: (_data, variables) => {
        void queryClient.invalidateQueries({
          queryKey: getAdminServiceListDeploymentsQueryKey(
            variables.org,
            variables.project,
            DEV_DEPLOYMENTS_PARAMS,
          ),
        });
      },
    },
  });
}

/**
 * Invalidates the dev deployments query, triggering a refetch.
 */
export function invalidateDevDeployments(org: string, project: string) {
  return queryClient.invalidateQueries({
    queryKey: getAdminServiceListDeploymentsQueryKey(
      org,
      project,
      DEV_DEPLOYMENTS_PARAMS,
    ),
  });
}

function isTransitionalStatus(status: V1DeploymentStatus | undefined): boolean {
  return (
    status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
    status === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING ||
    status === V1DeploymentStatus.DEPLOYMENT_STATUS_STOPPING ||
    status === V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING
  );
}

function isActiveDeployment(d: V1Deployment): boolean {
  return (
    d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
    d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
    d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING
  );
}
