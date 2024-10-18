import { checkUserAccess } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import { getPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
import { fetchOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors";
import { error, redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params: { organization }, url }) => {
  let redirectUrl = `/${organization}`;
  try {
    const issues = await fetchOrganizationBillingIssues(organization);
    const paymentIssues = getPaymentIssues(issues);
    if (paymentIssues.length) {
      redirectUrl = await fetchPaymentsPortalURL(
        organization,
        `${url.protocol}//${url.host}/${organization}`,
      );
    } else {
      redirectUrl = `/${organization}/-/settings/billing/callback`;
    }
  } catch (e) {
    if (e.response?.status !== 403 || (await checkUserAccess())) {
      throw error(e.response.status, "Error fetching billing issues");
    }
  }
  throw redirect(307, redirectUrl);
};
