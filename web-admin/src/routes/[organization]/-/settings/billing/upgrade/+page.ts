import { getPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import { createPaymentCheckoutSessionURL } from "@rilldata/web-admin/features/billing/plans/selectors";
import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({
  params: { organization },
  url,
  parent,
}) => {
  const { issues } = await parent();
  const paymentIssues = getPaymentIssues(issues);
  if (paymentIssues.length) {
    // Use Stripe Checkout for a better payment UX with multiple payment options
    const successUrl = `${url.protocol}//${url.host}/${organization}/-/upgrade-callback`;
    const cancelUrl = `${url.protocol}//${url.host}/${organization}/-/settings/billing`;
    throw redirect(
      307,
      await createPaymentCheckoutSessionURL(organization, successUrl, cancelUrl),
    );
  } else {
    throw redirect(307, `/${organization}/-/upgrade-callback`);
  }
};
