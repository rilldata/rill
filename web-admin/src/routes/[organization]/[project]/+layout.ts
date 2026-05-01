import { type RpcStatus } from "@rilldata/web-admin/client";
import { hasBlockerIssues } from "@rilldata/web-admin/features/billing/selectors";
import {
  branchPathPrefix,
  extractBranchFromPath,
} from "@rilldata/web-admin/features/branches/branch-utils";
import { fetchAllProjectsHibernating } from "@rilldata/web-admin/features/organizations/selectors";
import { error, redirect } from "@sveltejs/kit";
import { isAxiosError } from "axios";
import { maybeRedirectToEditableDeployment } from "@rilldata/web-admin/features/branches/deployment-utils.ts";
import { isEditPage } from "@rilldata/web-admin/features/navigation/nav-utils.ts";

// Sections hidden on branch views; visiting them redirects to the branch home.
const BRANCH_HIDDEN_SECTIONS = /\/-\/(alerts|reports|status|settings)(\/|$)/;

export const load = async ({
  params: { organization, project },
  parent,
  route,
  url,
}) => {
  const activeBranch = extractBranchFromPath(url.pathname);
  if (activeBranch && BRANCH_HIDDEN_SECTIONS.test(url.pathname)) {
    throw redirect(
      307,
      `/${organization}/${project}${branchPathPrefix(activeBranch)}`,
    );
  }

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

  if (!isEditPage({ route })) {
    await maybeRedirectToEditableDeployment(organization, project, url);
  }
};
