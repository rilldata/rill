import {
  adminServiceGetPaymentsPortalURL,
  adminServiceListPublicBillingPlans,
  getAdminServiceGetPaymentsPortalURLQueryKey,
  getAdminServiceListPublicBillingPlansQueryKey,
} from "@rilldata/web-admin/client";
import { isTeamPlan } from "@rilldata/web-admin/features/billing/plans/utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import type { Page } from "@sveltejs/kit";
import { DateTime } from "luxon";

export async function fetchTeamPlan() {
  const plansResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListPublicBillingPlansQueryKey(),
    queryFn: () => adminServiceListPublicBillingPlans(),
  });

  return plansResp.plans.find(isTeamPlan);
}

/**
 * We cannot prefetch this since the url is short-lived and single use for security purposes.
 * So we fetch just when we need it.
 */
export async function fetchPaymentsPortalURL(
  organization: string,
  returnUrl: string,
) {
  const portalUrlResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceGetPaymentsPortalURLQueryKey(organization, {
      returnUrl,
    }),
    queryFn: () =>
      adminServiceGetPaymentsPortalURL(organization, {
        returnUrl,
      }),
    // always refetch since the signed url will expire
    // TODO: figure out expiry time and use that instead
    cacheTime: 0,
    staleTime: 0,
  });

  return portalUrlResp.url;
}

export function getBillingUpgradeUrl(page: Page, organization: string) {
  return `${page.url.protocol}//${page.url.host}/${organization}/-/upgrade-callback`;
}

export function getNextBillingCycleDate(curEndDateRaw: string): string {
  const curEndDate = DateTime.fromJSDate(new Date(curEndDateRaw));
  if (!curEndDate.isValid) return "Unknown";
  const nextStartDate = curEndDate.plus({ day: 1 });
  return nextStartDate.toLocaleString(DateTime.DATE_MED);
}
