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
 * Returns the dev deployment for a specific branch (if any).
 * Used by the edit layout to find the deployment matching the `@branch` URL.
 */
export function useDevDeploymentByBranch(
  org: string,
  project: string,
  branch: string | undefined,
): Readable<{ data: V1Deployment | null; isLoading: boolean }> {
  const deploymentsQuery = useDevDeployments(org, project);

  return derived(deploymentsQuery, ($query) => {
    if ($query.isLoading) {
      return { data: null, isLoading: true };
    }
    if (!branch) {
      return { data: null, isLoading: false };
    }
    const found =
      $query.data?.deployments?.find((d) => d.branch === branch) ?? null;
    return { data: found, isLoading: false };
  });
}

/**
 * Mutation to create a dev deployment with editable=true.
 */
export function useCreateDevDeployment() {
  return createAdminServiceCreateDeployment({
    mutation: {
      onSuccess: (_data, variables) => {
        void invalidateDeployments(variables.org, variables.project);
      },
    },
  });
}

/**
 * Invalidates all deployment queries for a project, triggering a refetch.
 * Uses the base key (no params) so it matches both dev-scoped and
 * unscoped queries (e.g., BranchSelector).
 */
export function invalidateDeployments(org: string, project: string) {
  return queryClient.invalidateQueries({
    queryKey: getAdminServiceListDeploymentsQueryKey(org, project),
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

export function isActiveDeployment(d: V1Deployment): boolean {
  return (
    d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
    d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
    d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING
  );
}
