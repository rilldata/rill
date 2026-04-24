import {
  createAdminServiceListDeployments,
  getAdminServiceListDeploymentsQueryKey,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { derived } from "svelte/store";

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
 * Invalidates all deployment queries for a project, triggering a refetch.
 * Uses the base key (no params) so it matches both dev-scoped and
 * unscoped queries (e.g., BranchSelector).
 */
export function invalidateDeployments(org: string, project: string) {
  return queryClient.invalidateQueries({
    queryKey: getAdminServiceListDeploymentsQueryKey(org, project),
  });
}
