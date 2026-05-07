import {
  adminServiceListDeployments,
  getAdminServiceListDeploymentsQueryKey,
  V1DeploymentStatus,
  type V1Deployment,
  createAdminServiceListProjectsForOrganization,
  getAdminServiceListDeploymentsQueryOptions,
  type V1Project,
} from "@rilldata/web-admin/client";
import {
  extractBranchFromPath,
  injectBranchIntoPath,
} from "@rilldata/web-admin/features/branches/branch-utils.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { redirect } from "@sveltejs/kit";
import { derived, type Readable } from "svelte/store";
import { createQueries } from "@tanstack/svelte-query";

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

/**
 * If the project has no active prod deployment but does have an active
 * editable branch deployment, redirect into `/-/edit` for that branch.
 *
 * This is a convenience for direct links to unpublished projects so users
 * land somewhere usable instead of an empty/hibernating prod page. No-ops
 * when prod is healthy, when there's no editable branch to fall back to,
 * when the user is already on a branch URL, or on the `/-/deploying` or
 * `/-/invite` transitional screens.
 */
export async function maybeRedirectToEditableDeployment(
  organization: string,
  project: string,
  url: URL,
) {
  // The deploying and invite pages are transitional screens shown while a
  // prod deployment is still provisioning. Do not redirect away from them.
  if (
    url.pathname.endsWith("/-/deploying") ||
    url.pathname.endsWith("/-/invite")
  )
    return;

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
    injectBranchIntoPath(
      `/${organization}/${project}/-/edit`,
      editableDeployment.branch,
    ),
  );
}

export function getDeploymentsForProjectsInOrg(
  org: string,
): Readable<{ project: V1Project; deployments: V1Deployment[] }[]> {
  const projectsQuery = createAdminServiceListProjectsForOrganization(org);

  return derived(projectsQuery, (projectsResp, set) => {
    const projects = projectsResp.data?.projects ?? [];
    return createQueries(
      {
        queries: projects.map((project) =>
          getAdminServiceListDeploymentsQueryOptions(
            org,
            project.name ?? "",
            {},
            {},
          ),
        ),
        combine: (projectDeploymentsResponses) => {
          return projectDeploymentsResponses.map((projectDeployments, i) => ({
            project: projects[i],
            deployments: projectDeployments.data?.deployments ?? [],
          }));
        },
      },
      queryClient,
    ).subscribe(set);
  });
}
