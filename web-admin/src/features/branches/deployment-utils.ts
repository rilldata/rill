import {
  getAdminServiceListDeploymentsQueryKey,
  V1DeploymentStatus,
  type V1Deployment,
  adminServiceListDeployments,
} from "@rilldata/web-admin/client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { redirect } from "@sveltejs/kit";
import {
  branchPathPrefix,
  extractBranchFromPath,
} from "@rilldata/web-admin/features/branches/branch-utils.ts";

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
  // The deploying page is a transitional progress screen for a prod deployment
  // that is still provisioning. Do not redirect away from it.
  if (url.pathname.endsWith("/-/deploying")) return;

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
  if (!isActiveEditableDeployment) return;

  // If user is already in a specific deployment do not redirect.
  // This method is meant as a convenience for direct links to unpublished project.
  const currentBranch = extractBranchFromPath(url.pathname);
  if (currentBranch) return;

  throw redirect(
    307,
    `/${organization}/${project}${branchPathPrefix(editableDeployment.branch)}/-/edit`,
  );
}
