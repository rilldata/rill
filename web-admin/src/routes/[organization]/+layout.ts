import type { RpcStatus, V1BillingIssue } from "@rilldata/web-admin/client";
import { fetchOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
import { error } from "@sveltejs/kit";
import { isAxiosError } from "axios";

export const load = async ({ params: { organization }, parent }) => {
  const { user, organizationPermissions } = await parent();

  if (!organizationPermissions.readOrg) {
    throw error(403, "You do not have permission to access this organization");
  }

  let issues: V1BillingIssue[] = [];
  if (user) {
    // only try to get issues if the user can read org
    // also public projects will not have a user but will have `readOrg` permission
    try {
      issues = await fetchOrganizationBillingIssues(organization);
    } catch (e) {
      if (!isAxiosError<RpcStatus>(e) || !e.response) {
        throw error(500, "Error fetching billing issues");
      }

      throw error(e.response.status, e.response.data.message);
    }
  }

  return {
    organizationPermissions,
    issues,
  };
};
