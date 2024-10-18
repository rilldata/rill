import type { V1OrganizationPermissions } from "@rilldata/web-admin/client";
import { checkUserAccess } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import {
  fetchOrganizationBillingIssues,
  hasBlockerIssues,
} from "@rilldata/web-admin/features/billing/selectors";
import {
  fetchAllProjectsHibernating,
  fetchOrganizationPermissions,
} from "@rilldata/web-admin/features/organizations/selectors";
import { error, redirect } from "@sveltejs/kit";

export const load = async ({ params: { organization } }) => {
  let organizationPermissions: V1OrganizationPermissions = {};
  try {
    organizationPermissions = await fetchOrganizationPermissions(organization);
  } catch (e) {
    if (e.response?.status !== 403 || (await checkUserAccess())) {
      throw error(e.response.status, "Error fetching organization");
    }
  }
  if (!organizationPermissions.manageOrg) {
    return;
  }

  let shouldRedirectToProjectsList = false;

  try {
    const issues = await fetchOrganizationBillingIssues(organization);
    // if all projects were hibernated due to a blocker issue on org then take the user to projects page
    if (
      hasBlockerIssues(issues) &&
      (await fetchAllProjectsHibernating(organization))
    ) {
      shouldRedirectToProjectsList = true;
    }
  } catch (e) {
    console.error(e);
    if (e.response?.status !== 403 || (await checkUserAccess())) {
      throw error(e.response.status, "Error fetching billing issues");
    }
  }

  if (shouldRedirectToProjectsList) {
    throw redirect(307, `/${organization}`);
  }
};
