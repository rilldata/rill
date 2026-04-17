import {
  V1DeploymentStatus,
  getAdminServiceGetProjectQueryKey,
  getAdminServiceListDeploymentsQueryKey,
  type V1GetProjectResponse,
  type V1ListDeploymentsResponse,
} from "@rilldata/web-admin/client";
import { invalidateDeployments } from "./deployment-utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

/**
 * Optimistically set a deployment's status in both the ListDeployments
 * and GetProject caches, then mark both stale (without immediate refetch;
 * transient-status polling picks up the real state).
 */
export function optimisticallySetStatus(
  organization: string,
  project: string,
  deploymentId: string,
  branch: string | undefined,
  newStatus: V1DeploymentStatus,
) {
  const listKey = getAdminServiceListDeploymentsQueryKey(
    organization,
    project,
    {},
  );
  queryClient.setQueryData<V1ListDeploymentsResponse>(listKey, (old) => {
    if (!old?.deployments) return old;
    return {
      ...old,
      deployments: old.deployments.map((d) =>
        d.id === deploymentId ? { ...d, status: newStatus } : d,
      ),
    };
  });
  void queryClient.invalidateQueries({
    queryKey: getAdminServiceListDeploymentsQueryKey(organization, project),
    refetchType: "none",
  });

  const projectQueryKey = getAdminServiceGetProjectQueryKey(
    organization,
    project,
    branch ? { branch } : undefined,
  );
  queryClient.setQueryData<V1GetProjectResponse>(projectQueryKey, (old) => {
    if (!old?.deployment) return old;
    return {
      ...old,
      deployment: { ...old.deployment, status: newStatus },
    };
  });
  void queryClient.invalidateQueries({
    queryKey: getAdminServiceGetProjectQueryKey(organization, project),
    refetchType: "none",
  });
}

/**
 * Optimistically remove a deployment from the ListDeployments cache
 * and trigger a background refetch. Replaces the need for client-side
 * `deletedIds` tracking.
 */
export function optimisticallyRemoveDeployment(
  organization: string,
  project: string,
  deploymentId: string,
) {
  const listKey = getAdminServiceListDeploymentsQueryKey(
    organization,
    project,
    {},
  );
  queryClient.setQueryData<V1ListDeploymentsResponse>(listKey, (old) => {
    if (!old?.deployments) return old;
    return {
      ...old,
      deployments: old.deployments.filter((d) => d.id !== deploymentId),
    };
  });
  void invalidateDeployments(organization, project);
}
