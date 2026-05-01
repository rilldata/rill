import {
  getAdminServiceListDeploymentsQueryKey,
  V1DeploymentStatus,
  type V1Deployment,
  adminServiceListDeployments,
  createAdminServiceGetCurrentUser,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { redirect } from "@sveltejs/kit";
import {
  extractBranchFromPath,
  injectBranchIntoPath,
} from "@rilldata/web-admin/features/branches/branch-utils.ts";
import { useDevDeployments } from "@rilldata/web-admin/features/edit-session/use-edit-session.ts";
import { derived } from "svelte/store";

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

export async function maybeRedirectToEditableDeployment(
  organization: string,
  project: string,
  url: URL,
) {
  const deploymentsResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListDeploymentsQueryKey(organization, project, {}),
    queryFn: () => adminServiceListDeployments(organization, project, {}),
    // Do not refetch in this loader function. Refetch strategy is instead managed in BranchesSection.svelte
    staleTime: Infinity,
  });
  const prodDeployment = deploymentsResp.deployments?.find(isProdDeployment);
  const editableDeployment = deploymentsResp.deployments?.find(
    (d) => d.editable,
  );

  // There is an active prod deployment, so no need for redirect.
  const isActiveProdDeployment =
    prodDeployment && isActiveDeployment(prodDeployment);
  if (isActiveProdDeployment || !editableDeployment?.branch) return;

  const isActiveEditableDeployment =
    editableDeployment && isActiveDeployment(editableDeployment);
  // Editable deployment is inactive as well, project is probably hibernating, skip redirect.
  if (!isActiveEditableDeployment && prodDeployment) return;

  // If user is already in a specific deployment do not redirect.
  // This method is meant as a convenience for direct links to unpublished project.
  const currentBranch = extractBranchFromPath(url.pathname);
  if (currentBranch) return;

  throw redirect(
    307,
    injectBranchIntoPath(
      `/${organization}/${project}`,
      editableDeployment.branch,
    ),
  );
}

export function getSingleEditableDeploymentHref(
  organization: string,
  project: string,
) {
  const userQuery = createAdminServiceGetCurrentUser();
  const devDeploymentsQuery = useDevDeployments(organization, project);
  return derived(
    [userQuery, devDeploymentsQuery],
    ([userResp, devDeploymentsResp]) => {
      const currentUserId = userResp.data?.user?.id;
      const activeDeploymentsForUser =
        devDeploymentsResp.data?.deployments?.filter(
          (d) => d.ownerUserId === currentUserId && isActiveDeployment(d),
        );
      if (activeDeploymentsForUser?.length !== 1) return undefined;

      const singleActiveDeployment = activeDeploymentsForUser[0];
      if (!singleActiveDeployment?.branch) return undefined;
      return injectBranchIntoPath(
        `/${organization}/${project}/-/edit`,
        singleActiveDeployment.branch,
      );
    },
  );
}
