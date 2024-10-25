import type { V1BillingIssue } from "@rilldata/web-admin/client";
import { redirectToLoginIfNotLoggedIn } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import { fetchOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { organization }, parent }) => {
  const { organizationPermissions } = await parent();

  let issues: V1BillingIssue[] = [];
  try {
    issues = await fetchOrganizationBillingIssues(organization);
  } catch (e) {
    if (e.response?.status !== 403 || (await redirectToLoginIfNotLoggedIn())) {
      throw error(e.response.status, "Error fetching billing issues");
    }
  }

  return {
    organizationPermissions,
    issues,
  };
};
