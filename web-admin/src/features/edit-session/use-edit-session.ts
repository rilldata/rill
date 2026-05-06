import {
  createAdminServiceListDeployments,
  getAdminServiceGetCurrentUserQueryOptions,
  getAdminServiceListDeploymentsQueryKey,
  getAdminServiceListDeploymentsQueryOptions,
  V1DeploymentStatus,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { createQueries } from "@tanstack/svelte-query";
import { derived } from "svelte/store";

/**
 * Lists all deployments for a project (no polling).
 *
 * Uses an empty params object (`{}`) so the TanStack Query cache key matches
 * the BranchesSection query. This avoids duplicate ListDeployments requests
 * when both consumers are mounted on the same page; callers filter to dev
 * deployments client-side.
 *
 * Freshness is maintained by invalidateDeployments() after create/delete
 * mutations and the BranchesSection's transitory-status polling.
 */
export function useAllDeployments(org: string, project: string) {
  return createAdminServiceListDeployments(org, project, {});
}

/**
 * Lists dev deployments for a project. Shares the same underlying query as
 * useAllDeployments to avoid duplicate network requests.
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
 * Lists the current user's editable dev deployments for a project. Combines
 * the deployment list and current-user queries so callers get a single store
 * with the filtered, sorted slice plus the current user's id.
 *
 * Stopped/errored deployments are included so the user can resume or retry;
 * deleting/deleted deployments are filtered out.
 */
export function useOwnDevDeployments(org: string, project: string) {
  return createQueries({
    queries: [
      getAdminServiceListDeploymentsQueryOptions(org, project, {}),
      getAdminServiceGetCurrentUserQueryOptions(),
    ],
    combine: ([deploymentsQuery, userQuery]) => {
      const currentUserId = userQuery.data?.user?.id;
      const all = deploymentsQuery.data?.deployments ?? [];
      const ownDeployments: V1Deployment[] = all
        .filter(
          (d) =>
            d.environment === "dev" &&
            d.ownerUserId === currentUserId &&
            d.editable &&
            d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETING &&
            d.status !== V1DeploymentStatus.DEPLOYMENT_STATUS_DELETED,
        )
        .sort((a, b) => (b.updatedOn ?? "").localeCompare(a.updatedOn ?? ""));
      return {
        ownDeployments,
        currentUserId,
        isLoading: deploymentsQuery.isLoading || userQuery.isLoading,
      };
    },
  });
}

/**
 * Invalidates all deployment queries for a project, triggering a refetch.
 * Uses the base key (no params) so it matches both dev-scoped and
 * unscoped queries.
 */
export function invalidateDeployments(org: string, project: string) {
  return queryClient.invalidateQueries({
    queryKey: getAdminServiceListDeploymentsQueryKey(org, project),
  });
}
