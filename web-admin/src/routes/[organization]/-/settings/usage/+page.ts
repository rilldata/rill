import { isEnterprisePlan } from "@rilldata/web-admin/features/billing/plans/utils";
import { error, redirect } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params: { organization }, parent }) => {
  const { subscription, billingPortalUrl } = await parent();

  if (!billingPortalUrl) {
    throw redirect(307, `/${organization}/-/settings`);
  }

  // Orgs on an Enterprise Plan should not see this page
  if (subscription?.plan && isEnterprisePlan(subscription.plan.planType)) {
    throw error(404, "Page not found");
  }
};
