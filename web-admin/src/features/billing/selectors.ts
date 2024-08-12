import {
  createAdminServiceGetBillingSubscription,
  createAdminServiceListPublicBillingPlans,
} from "@rilldata/web-admin/client";
import { derived } from "svelte/store";

export function getPlanForOrg(org: string) {
  return derived(
    [
      createAdminServiceGetBillingSubscription(org),
      createAdminServiceListPublicBillingPlans(),
    ],
    ([subscriptionResp, plansResp]) => {
      if (!subscriptionResp.data?.subscription || !plansResp.data?.plans) {
        return undefined;
      }

      return plansResp.data.plans.find(
        (plan) => plan.id === subscriptionResp.data.subscription.planId,
      );
    },
  );
}
