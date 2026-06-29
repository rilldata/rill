import { createQuery } from "@tanstack/svelte-query";
import {
  adminServiceGetPaymentsPortalURL,
  adminServiceListPublicBillingPlans,
  createAdminServiceGetBillingCreditBalance,
  createAdminServiceGetBillingProjectCredentials,
  getAdminServiceGetPaymentsPortalURLQueryKey,
  getAdminServiceListPublicBillingPlansQueryKey,
  type V1BillingIssue,
  V1BillingPlanType,
  type V1Subscription,
} from "@rilldata/web-admin/client";
import {
  getDeploymentsForProjectsInOrg,
  isActiveDeployment,
  isProdDeployment,
} from "@rilldata/web-admin/features/branches/deployment-utils";
import {
  isEnterprisePlan,
  isFreePlan,
  isGrowthPlan,
  isManagedPlan,
  isProPlan,
  isStarterPlan,
  isTeamPlan,
} from "@rilldata/web-admin/features/billing/plans/utils";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import * as m from "@rilldata/web-common/paraglide/messages.js";
import type { Page } from "@sveltejs/kit";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { DateTime } from "luxon";
import { derived, type Readable } from "svelte/store";
import type { PlanTier } from "@rilldata/web-admin/features/billing/plans/types.ts";
import type { CategorisedOrganizationBillingIssues } from "@rilldata/web-admin/features/billing/selectors.ts";
import { SELF_SERVE_PLANS_BY_NAME } from "@rilldata/web-admin/features/billing/plans/plan-details.ts";

export async function maybeFetchPublicPlanByName(planName: string) {
  const staticPlan = SELF_SERVE_PLANS_BY_NAME[planName];
  if (staticPlan) return staticPlan;

  const plansResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceListPublicBillingPlansQueryKey(),
    queryFn: () => adminServiceListPublicBillingPlans(),
  });

  const remotePlan = plansResp.plans?.find((p) => p.name === planName);
  if (remotePlan) return remotePlan;

  throw new Error(`Plan ${planName} not found`);
}

/**
 * We cannot prefetch this since the url is short-lived and single use for security purposes.
 * So we fetch just when we need it.
 */
export async function fetchPaymentsPortalURL(
  organization: string,
  returnUrl: string,
  setup?: boolean,
) {
  const portalUrlResp = await queryClient.fetchQuery({
    queryKey: getAdminServiceGetPaymentsPortalURLQueryKey(organization, {
      returnUrl,
      setup,
    }),
    queryFn: () =>
      adminServiceGetPaymentsPortalURL(organization, {
        returnUrl,
        setup,
      }),
    // always refetch since the signed url will expire
    // TODO: figure out expiry time and use that instead
    gcTime: 0,
    staleTime: 0,
  });

  return portalUrlResp.url ?? "";
}

export function getBillingUpgradeUrl(
  page: Page,
  organization: string,
  planName?: string,
) {
  const url = new URL(page.url);
  url.pathname = `/${organization}/-/upgrade-callback`;
  if (planName) {
    url.searchParams.set("plan", planName);
  }
  return url.toString();
}

export function getNextBillingCycleDate(curEndDateRaw: string): string {
  const curEndDate = DateTime.fromJSDate(new Date(curEndDateRaw));
  if (!curEndDate.isValid) return m.billing_unknown();
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
  return fetch(url.toString(), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${accessToken}`,
    },
  }).then((r) => r.json());
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

// Daily run-rate from the current configuration (units × list rate × 24h).
// Used as a placeholder until the billing usage API exposes accrued $ values.
const RATE_PER_UNIT_HR = 0.15;
const HOURS_PER_DAY = 24;

export type BillingStats = {
  prodSlots: number;
  devSlots: number;
  prodDailyCost: number;
  devDailyCost: number;
};

export function getBillingStatsForOrg(org: string): Readable<BillingStats> {
  return derived(getDeploymentsForProjectsInOrg(org), (projectDeployments) => {
    // Compute units = project.{prod,dev}Slots × number of running {prod,dev}
    // deployments for that project, summed across the org.
    const items = projectDeployments ?? [];
    const prodSlots = items.reduce((sum, { project, deployments }) => {
      const slots = Number(project.prodSlots ?? 0);
      const running = deployments.filter(
        (d) => isProdDeployment(d) && isActiveDeployment(d),
      ).length;
      return sum + slots * running;
    }, 0);
    const devSlots = items.reduce((sum, { project, deployments }) => {
      const slots = Number(project.devSlots ?? 0);
      const running = deployments.filter(
        (d) => !isProdDeployment(d) && isActiveDeployment(d),
      ).length;
      return sum + slots * running;
    }, 0);

    // Daily run-rate from the current configuration (units × list rate × 24h).
    // Used as a placeholder until the billing usage API exposes accrued $ values.
    const prodDailyCost = prodSlots * RATE_PER_UNIT_HR * HOURS_PER_DAY;
    const devDailyCost = devSlots * RATE_PER_UNIT_HR * HOURS_PER_DAY;

    return {
      prodSlots,
      devSlots,
      prodDailyCost,
      devDailyCost,
    };
  });
}

export function getPlanTierForSubscription(
  subscription: V1Subscription | undefined,
  categorisedIssues: CategorisedOrganizationBillingIssues | undefined,
): PlanTier {
  if (!subscription) {
    return categorisedIssues?.trial ? "free" : "pro";
  }

  const plan = subscription?.plan;
  const planType = plan?.planType;
  const planName = plan?.name ?? "";
  // Prefer planType enum when available; fall back to plan.name string matching
  if (
    planType === V1BillingPlanType.BILLING_PLAN_TYPE_TEAM ||
    isTeamPlan(planName)
  )
    return "team";
  if (
    planType === V1BillingPlanType.BILLING_PLAN_TYPE_MANAGED ||
    isManagedPlan(planName)
  )
    return "managed";
  if (
    planType === V1BillingPlanType.BILLING_PLAN_TYPE_ENTERPRISE ||
    isEnterprisePlan(planName)
  )
    return "enterprise";
  if (
    planType === V1BillingPlanType.BILLING_PLAN_TYPE_PRO ||
    isProPlan(planName)
  )
    return "pro";
  if (
    planType === V1BillingPlanType.BILLING_PLAN_TYPE_STARTER ||
    isStarterPlan(planName)
  )
    return "starter";
  if (
    planType === V1BillingPlanType.BILLING_PLAN_TYPE_GROWTH ||
    isGrowthPlan(planName)
  )
    return "growth";
  if (isFreePlan(planName)) return "free";
  // free_trial, no plan, cancelled — all trial
  return "trial";
}

export function getBillingCycleDates(subscription: V1Subscription | undefined) {
  const periodStart = subscription?.currentBillingCycleStartDate
    ? DateTime.fromISO(subscription.currentBillingCycleStartDate)
    : DateTime.now().startOf("month");
  const formattedPeriodStart = periodStart.toLocaleString({
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  const periodEnd = subscription?.currentBillingCycleEndDate
    ? DateTime.fromISO(subscription.currentBillingCycleEndDate).minus({
        days: 1,
      })
    : periodStart.endOf("month");
  const formattedPeriodEnd = periodEnd.toLocaleString({
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  const dueDate = subscription?.currentBillingCycleEndDate
    ? DateTime.fromISO(subscription.currentBillingCycleEndDate)
    : periodStart.plus({ months: 1 });
  const formattedDueDate = dueDate.toLocaleString({
    month: "short",
    day: "numeric",
    year: "numeric",
  });

  return {
    formattedPeriodStart,
    formattedPeriodEnd,
    formattedDueDate,
  };
}

const TOTAL_CREDIT = 250;

export function getPlanCredits(
  org: string,
  billingIssue: V1BillingIssue | undefined,
) {
  return derived(
    createAdminServiceGetBillingCreditBalance(org, {}),
    (creditsBalance) => {
      if (!creditsBalance.data) {
        // Set everything to 0 while the data is potentially loading.
        return { usedCredit: 0, availableCredit: 0, creditPercent: 0 };
      }

      const totalCredit =
        billingIssue?.metadata?.onCreditTrial?.creditAllocation ?? TOTAL_CREDIT;

      const availableCredit = creditsBalance.data?.balance ?? totalCredit;
      const usedCredit = totalCredit - availableCredit;
      const creditPercent = Math.round((usedCredit / totalCredit) * 100);
      return {
        usedCredit,
        availableCredit,
        creditPercent,
      };
    },
  );
}
