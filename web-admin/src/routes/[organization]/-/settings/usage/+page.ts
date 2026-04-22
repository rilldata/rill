import {
  isFreePlan,
  isProPlan,
  isTrialPlan,
} from "@rilldata/web-admin/features/billing/plans/utils";
import { error } from "@sveltejs/kit";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params: { organization }, parent }) => {
  const { subscription } = await parent();
  const planName = subscription?.plan?.name ?? "";

  const allowed =
    isFreePlan(planName) || isProPlan(planName) || isTrialPlan(planName);

  if (planName && !allowed) {
    throw error(404, "Page not found");
  }

  return { organization };
};
