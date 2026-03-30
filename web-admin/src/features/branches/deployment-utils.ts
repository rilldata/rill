import {
  getAdminServiceListDeploymentsQueryKey,
  V1DeploymentStatus,
  type V1Deployment,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

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

/** Canonical check: a deployment is production if its environment is "prod". */
export function isProdDeployment(d: V1Deployment): boolean {
  return d.environment === "prod";
}

/**
 * Deduplicate deployments by branch, keeping only the most recently
 * updated deployment per branch. Optionally filters out deployments
 * matching a predicate (e.g., deleted ones).
 */
export function deduplicateDeployments(
  deployments: V1Deployment[],
  exclude?: (d: V1Deployment) => boolean,
): V1Deployment[] {
  const byBranch = new Map<string, V1Deployment>();
  for (const d of deployments) {
    if (exclude?.(d)) continue;
    const key = d.branch ?? "";
    const existing = byBranch.get(key);
    // updatedOn is an ISO 8601 timestamp; lexicographic comparison is correct.
    if (!existing || (d.updatedOn ?? "") > (existing.updatedOn ?? "")) {
      byBranch.set(key, d);
    }
  }
  return [...byBranch.values()];
}
