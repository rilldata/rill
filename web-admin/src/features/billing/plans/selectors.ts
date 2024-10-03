import {
  createAdminServiceListPublicBillingPlans,
  type V1BillingPlan,
} from "@rilldata/web-admin/client";
import {
  isTeamPlan,
  isTrialPlan,
} from "@rilldata/web-admin/features/billing/plans/utils";
import { DateTime } from "luxon";

export function getCategorisedPlans() {
  return createAdminServiceListPublicBillingPlans({
    query: {
      select: (data) => {
        let trialPlan: V1BillingPlan;
        let teamPlan: V1BillingPlan;

        data.plans.forEach((p) => {
          if (isTrialPlan(p)) {
            trialPlan = p;
          } else if (isTeamPlan(p) && !teamPlan) {
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

export function getNextBillingCycleDate(curEndDateRaw: string): string {
  const curEndDate = DateTime.fromJSDate(new Date(curEndDateRaw));
  if (!curEndDate.isValid) return "Unknown";
  const nextStartDate = curEndDate.plus({ day: 1 });
  return nextStartDate.toLocaleString(DateTime.DATE_MED);
}
