import { createQuery } from "@tanstack/svelte-query";
import {
  adminServiceGetPaymentsPortalURL,
  adminServiceListPublicBillingPlans,
  createAdminServiceGetBillingProjectCredentials,
  getAdminServiceGetPaymentsPortalURLQueryKey,
  getAdminServiceListPublicBillingPlansQueryKey,
} from "@rilldata/web-admin/client";
import { isTeamPlan } from "@rilldata/web-admin/features/billing/plans/utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { fetchWrapper } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { Page } from "@sveltejs/kit";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { DateTime } from "luxon";
import { derived } from "svelte/store";

export async function fetchTeamPlan() {
  const plansResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListPublicBillingPlansQueryKey(),
    queryFn: () => adminServiceListPublicBillingPlans(),
  });

  return plansResp.plans?.find((p) => isTeamPlan(p.name ?? ""));
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
    gcTime: 0,
    staleTime: 0,
  });

  return portalUrlResp.url ?? "";
}

export function getBillingUpgradeUrl(page: Page, organization: string) {
  const url = new URL(page.url);
  url.pathname = `/${organization}/-/upgrade-callback`;
  return url.toString();
}

export function getNextBillingCycleDate(curEndDateRaw: string): string {
  const curEndDate = DateTime.fromJSDate(new Date(curEndDateRaw));
  if (!curEndDate.isValid) return "Unknown";
  return curEndDate.toLocaleString(DateTime.DATE_MED);
}

export function getOrganizationUsageMetrics(
  organization: string,
): CreateQueryResult<UsageMetricsResponse> {
  return derived(
    [
      createAdminServiceGetBillingProjectCredentials({
        org: organization,
      }),
    ],
    ([credsResp], set) => {
      if (!credsResp.data) return;
      return getUsageMetrics(
        credsResp.data.runtimeHost ?? "",
        credsResp.data.instanceId ?? "",
        credsResp.data.accessToken ?? "",
      ).subscribe(set);
    },
  );
}

export type UsageMetricsResponse = {
  project_name: string;
  size: number;
}[];
function usageMetrics(
  runtimeHost: string,
  instanceId: string,
  accessToken: string,
): Promise<UsageMetricsResponse> {
  const url = new URL(runtimeHost);
  url.pathname = `/v1/instances/${instanceId}/api/usage-meter`;
  return fetchWrapper({
    url: url.toString(),
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
  });
}
export function getUsageMetrics(
  runtimeHost: string,
  instanceId: string,
  accessToken: string,
) {
  return createQuery(
    {
      queryKey: [
        `/v1/instances/${instanceId}/api/usage-meter`,
        runtimeHost,
        accessToken,
      ],
      queryFn: () => usageMetrics(runtimeHost, instanceId, accessToken),
    },
    queryClient,
  );
}
