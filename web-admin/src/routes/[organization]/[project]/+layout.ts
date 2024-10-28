import { hasBlockerIssues } from "@rilldata/web-admin/features/billing/selectors";
import { fetchAllProjectsHibernating } from "@rilldata/web-admin/features/organizations/selectors";
import { error, redirect } from "@sveltejs/kit";

export const load = async ({ params: { organization }, parent }) => {
  const { organizationPermissions, issues } = await parent();
  if (!organizationPermissions.manageOrg) {
    return;
  }

  let shouldRedirectToProjectsList = false;

  try {
    // if all projects were hibernated due to a blocker issue on org then take the user to projects page
    if (
      hasBlockerIssues(issues) &&
      (await fetchAllProjectsHibernating(organization))
    ) {
      shouldRedirectToProjectsList = true;
    }
  } catch (e) {
    console.error(e);
    throw error(e.response.status, "Error fetching billing issues");
  }

  if (shouldRedirectToProjectsList) {
    throw redirect(307, `/${organization}`);
  }
};
