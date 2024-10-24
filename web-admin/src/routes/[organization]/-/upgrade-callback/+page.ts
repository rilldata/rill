import type { V1BillingIssue } from "@rilldata/web-admin/client";
import { checkUserAccess } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import { getCancelledIssue } from "@rilldata/web-admin/features/billing/issues/getMessageForCancelledIssue";
import { getPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import { fetchOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params: { organization } }) => {
  let issues: V1BillingIssue[] = [];
  try {
    issues = await fetchOrganizationBillingIssues(organization);
  } catch (e) {
    if (e.response?.status !== 403 || (await checkUserAccess())) {
      throw error(e.response.status, "Error fetching billing issues");
    }
  }

  return {
    paymentIssues: getPaymentIssues(issues),
    cancelled: !!getCancelledIssue(issues),
  };
};
