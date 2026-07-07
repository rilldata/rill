import { redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";
import { PaidPlanTypes } from "@rilldata/web-admin/features/billing/plans/utils.ts";

export const load: PageLoad = async ({ params: { organization }, parent }) => {
  const { billingPortalUrl, subscription } = await parent();

  if (
    !billingPortalUrl ||
    (subscription?.plan?.planType && !PaidPlanTypes[subscription.plan.planType])
  ) {
    throw redirect(307, `/${organization}/-/settings`);
  }
};
