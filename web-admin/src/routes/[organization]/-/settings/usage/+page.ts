import { isEnterprisePlan } from "@rilldata/web-admin/features/billing/plans/utils";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ parent }) => {
  const { subscription } = await parent();

  // Orgs on an Enterprise Plan should not see this page
  if (subscription?.plan && isEnterprisePlan(subscription.plan)) {
    throw error(404, "Page not found");
  }
};
