import {
  adminServiceListPublicBillingPlans,
  getAdminServiceListPublicBillingPlansQueryKey,
  type RpcStatus,
} from "@rilldata/web-admin/client";
import { isTeamPlan } from "@rilldata/web-admin/features/billing/plans/utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { error } from "@sveltejs/kit";
import { isAxiosError } from "axios";
import type { PageLoad } from "./$types";

export const load: PageLoad = async ({ params: { organization }, parent }) => {
  // Get billing issues from parent layout
  const { issues } = await parent();

  // Fetch the team plan details
  try {
    const plansResp = await queryClient.fetchQuery({
      queryKey: getAdminServiceListPublicBillingPlansQueryKey(),
      queryFn: () => adminServiceListPublicBillingPlans(),
    });

    const teamPlan = plansResp.plans?.find((p) => isTeamPlan(p.name ?? ""));

    return {
      organization,
      teamPlan,
      issues,
    };
  } catch (e) {
    if (!isAxiosError<RpcStatus>(e) || !e.response) {
      throw error(500, "Error loading payment page");
    }

    throw error(e.response.status, e.response.data.message);
  }
};
