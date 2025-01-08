import type { V1BillingIssue } from "@rilldata/web-admin/client";
import { fetchOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
import { error } from "@sveltejs/kit";

export const load = async ({ params: { organization }, parent }) => {
  const { organizationPermissions, user } = await parent();

  let issues: V1BillingIssue[] = [];
  if (organizationPermissions.readOrg && user) {
    // only try to get issues if the user can read org
    // also public projects will not have a user but will have `readOrg` permission
    try {
      issues = await fetchOrganizationBillingIssues(organization);
    } catch (e) {
      if (e.response?.status !== 403) {
        throw error(e.response.status, "Error fetching billing issues");
      }
    }
  }

  return {
    organizationPermissions,
    issues,
  };
};
