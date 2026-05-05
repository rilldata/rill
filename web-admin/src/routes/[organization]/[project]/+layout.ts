import { type RpcStatus } from "@rilldata/web-admin/client";
import { hasBlockerIssues } from "@rilldata/web-admin/features/billing/selectors";
import { extractBranchFromPath } from "@rilldata/web-admin/features/branches/branch-utils.ts";
import { maybeRedirectToEditableDeployment } from "@rilldata/web-admin/features/branches/deployment-utils.ts";
import { isEditPage } from "@rilldata/web-admin/features/navigation/nav-utils.ts";
import { fetchAllProjectsHibernating } from "@rilldata/web-admin/features/organizations/selectors";
import { error, redirect } from "@sveltejs/kit";
import { isAxiosError } from "axios";

export const load = async ({
  params: { organization, project },
  parent,
  route,
  url,
}) => {
  const { organizationPermissions, issues } = await parent();

  if (!organizationPermissions.manageOrg) return;

  let projectHibernating = false;
  try {
    projectHibernating = await fetchAllProjectsHibernating(organization);
  } catch (e) {
    if (!isAxiosError<RpcStatus>(e) || !e.response) {
      throw error(500, "Error fetching projects for the organization");
    }

    throw error(e.response.status, e.response.data.message);
  }

  // if all projects were hibernated due to a blocker issue on org then take the user to projects page
  if (hasBlockerIssues(issues) && projectHibernating) {
    throw redirect(307, `/${organization}`);
  }

  // Edit pages handle their own branch routing; everything below is non-edit only.
  if (isEditPage({ route })) return;

  // Branch deployments are only viewable from inside `/-/edit`.
  const branch = extractBranchFromPath(url.pathname);
  if (branch) {
    throw error(404, "Branch deployments are only available from the editor.");
  }

  await maybeRedirectToEditableDeployment(organization, project, url);
};
