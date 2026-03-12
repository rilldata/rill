import {
  createAdminServiceCreateDeployment,
  createAdminServiceListDeployments,
  getAdminServiceListDeploymentsQueryKey,
  V1DeploymentStatus,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { derived, type Readable } from "svelte/store";

/**
 * Lists all deployments for a project (no polling).
 *
 * Uses an empty params object (`{}`) so the TanStack Query cache key matches
 * the BranchSelector's query. This avoids duplicate ListDeployments requests
 * when both components are mounted on the same page; callers filter to dev
 * deployments client-side.
 *
 * Freshness is maintained by:
 * - BranchSelector polling at 2s while its dropdown is open
 * - invalidateDeployments() called after create/delete mutations
 */
export function useAllDeployments(org: string, project: string) {
  return createAdminServiceListDeployments(org, project, {});
}

/**
 * Lists dev deployments for a project. Shares the same underlying query as
 * useAllDeployments (and BranchSelector) to avoid duplicate network requests.
 */
export function useDevDeployments(org: string, project: string) {
  const allQuery = useAllDeployments(org, project);
  return derived(allQuery, ($query) => ({
    ...$query,
    data: $query.data
      ? {
          ...$query.data,
          deployments: $query.data.deployments?.filter(
            (d) => d.environment === "dev",
          ),
        }
      : $query.data,
  }));
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

export function isActiveDeployment(d: V1Deployment): boolean {
  return (
    d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_PENDING ||
    d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_RUNNING ||
    d.status === V1DeploymentStatus.DEPLOYMENT_STATUS_UPDATING
  );
}
