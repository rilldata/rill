import {
  adminServiceListPublicBillingPlans,
  createAdminServiceListPublicBillingPlans,
  getAdminServiceListPublicBillingPlansQueryKey,
  type V1BillingPlan,
} from "@rilldata/web-admin/client";
import {
  isTeamPlan,
  isTrialPlan,
} from "@rilldata/web-admin/features/billing/plans/utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
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

export async function fetchTeamPlan() {
  const plansResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListPublicBillingPlansQueryKey(),
    queryFn: () => adminServiceListPublicBillingPlans(),
  });
  for (const plan of plansResp.plans) {
    if (isTeamPlan(plan)) return plan;
  }
  return undefined;
}

export function getNextBillingCycleDate(curEndDateRaw: string): string {
  const curEndDate = DateTime.fromJSDate(new Date(curEndDateRaw));
  if (!curEndDate.isValid) return "Unknown";
  const nextStartDate = curEndDate.plus({ day: 1 });
  return nextStartDate.toLocaleString(DateTime.DATE_MED);
}
