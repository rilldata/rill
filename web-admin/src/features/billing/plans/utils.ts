import { V1BillingPlanType } from "@rilldata/web-admin/client";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
import { DateTime } from "luxon";
import { writable } from "svelte/store";

export function formatUsageVsQuota(
  usageInBytes: number,
  storageLimitBytesPerDeployment: string,
): string {
  const quota = Number(storageLimitBytesPerDeployment);
  if (Number.isNaN(quota) || storageLimitBytesPerDeployment === "-1") return "";
  const formattedUsage = formatMemorySize(usageInBytes);
  const formattedQuota = formatMemorySize(quota);
  const percent =
    usageInBytes > quota
      ? "100+"
      : Math.round((usageInBytes * 100) / quota) + "";
  return `${formattedUsage} of ${formattedQuota} (${percent}%)`;
}

export function isTrialPlan(planType: V1BillingPlanType) {
  return planType === V1BillingPlanType.BILLING_PLAN_TYPE_TRIAL;
}

export function isTeamPlan(planType: V1BillingPlanType) {
  return planType === V1BillingPlanType.BILLING_PLAN_TYPE_TEAM;
}

export function isManagedPlan(planType: V1BillingPlanType) {
  return planType === V1BillingPlanType.BILLING_PLAN_TYPE_MANAGED;
}

export function isEnterprisePlan(planType: V1BillingPlanType) {
  return (
    !isTrialPlan(planType) && !isTeamPlan(planType) && !isManagedPlan(planType)
  );
}

export function getSubscriptionResumedText(endDate: string) {
  const date = DateTime.fromJSDate(new Date(endDate));
  if (!date.isValid || date.toMillis() < Date.now()) {
    return "today";
  }
  const resumeDate = date.plus({ day: 1 });
  return "on " + resumeDate.toLocaleString(DateTime.DATE_MED);
}

// Since this could be triggered in a route that could be navigated from,
// we add a global and show it in org route's layout
export const showWelcomeToRillDialog = writable(false);
