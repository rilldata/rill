import { getPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({
  params: { organization },
  parent,
}) => {
  const { issues } = await parent();
  const paymentIssues = getPaymentIssues(issues);
  if (paymentIssues.length) {
    // Redirect to the payment page which shows pricing and then opens Stripe Checkout
    throw redirect(307, `/${organization}/-/settings/billing/payment`);
  } else {
    throw redirect(307, `/${organization}/-/upgrade-callback`);
  }
};
