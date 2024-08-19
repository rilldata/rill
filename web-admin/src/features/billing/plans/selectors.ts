import {
  createAdminServiceListPublicBillingPlans,
  type V1BillingPlan,
} from "@rilldata/web-admin/client";

export function getCategorisedPlans() {
  return createAdminServiceListPublicBillingPlans({
    query: {
      select: (data) => {
        let trialPlan: V1BillingPlan;
        let teamPlan: V1BillingPlan;

        data.plans.forEach((p) => {
          if (p.default && p.trialPeriodDays) {
            trialPlan = p;
          } else if (p.name?.includes("Team") && !teamPlan) {
            teamPlan = p;
          }
        });

        return {
          trialPlan,
          teamPlan,
        };
      },
    },
  });
}
