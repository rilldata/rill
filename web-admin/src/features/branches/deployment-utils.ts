import {
  getAdminServiceListDeploymentsQueryKey,
  V1DeploymentStatus,
  type V1Deployment,
  adminServiceListDeployments,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { redirect } from "@sveltejs/kit";
import { injectBranchIntoPath } from "@rilldata/web-admin/features/branches/branch-utils.ts";
import { isProjectWelcomePage } from "@rilldata/web-admin/features/navigation/nav-utils.ts";

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

  const isActiveProdDeployment =
    prodDeployment && isActiveDeployment(prodDeployment);
  if (isActiveProdDeployment || !editableDeployment) return;

  throw redirect(
    307,
    injectBranchIntoPath(
      `/${organization}/${project}/-/edit`,
      editableDeployment.branch,
    ),
  );
}
