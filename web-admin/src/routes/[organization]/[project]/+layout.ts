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

export const load = async ({
  url,
  params: { organization, project },
  parent,
  route,
  url,
}) => {
  // Branches are accessible only via the editor surface. Any `/@branch`
  // URL outside of `/-/edit` (and outside of public `/-/share/` magic
  // links) is funneled into the equivalent edit route so users can never
  // see a branch in the production cloud view.
  const branch = extractBranchFromPath(url.pathname);
  if (branch) {
    const prefix = `/${organization}/${project}${branchPathPrefix(branch)}`;
    const subpath = url.pathname.slice(prefix.length);
    const onEdit = subpath.startsWith("/-/edit");
    const onShare = subpath.startsWith("/-/share/");
    if (!onEdit && !onShare) {
      let editSubpath: string;
      if (!subpath || subpath === "/") {
        editSubpath = "/-/edit/dashboards";
      } else if (
        subpath.startsWith("/explore/") ||
        subpath.startsWith("/canvas/")
      ) {
        editSubpath = "/-/edit" + subpath;
      } else {
        // Sub-pages without an editor equivalent (reports, alerts,
        // settings, status, etc.) collapse to the preview home.
        editSubpath = "/-/edit/dashboards";
      }
      // `url.hash` is unavailable in SvelteKit `load` (client-only); the
      // browser preserves the hash across the redirect on its own.
      throw redirect(303, prefix + editSubpath + url.search);
    }
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
