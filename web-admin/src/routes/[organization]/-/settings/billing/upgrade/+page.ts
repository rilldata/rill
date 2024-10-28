import { getPaymentIssues } from "@rilldata/web-admin/features/billing/issues/getMessageForPaymentIssues";
import { fetchPaymentsPortalURL } from "@rilldata/web-admin/features/billing/plans/selectors";
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
    throw redirect(
      307,
      await fetchPaymentsPortalURL(
        organization,
        `${url.protocol}//${url.host}/${organization}`,
      ),
    );
  } else {
    throw redirect(307, `/${organization}/-/settings/billing/callback`);
  }
};
