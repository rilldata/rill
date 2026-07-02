import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { DateTime } from "luxon";
import { writable } from "svelte/store";
import { V1BillingPlanType } from "@rilldata/web-admin/client";

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
  return m.billing_usage_of_quota({ usage: formattedUsage, quota: formattedQuota, percent });
}

// Mapping of externalID/planName to a type.
// Used in deciding banner message and to show different billing module in frontend.
// Make sure to update admin/billing/orb.go::getPlanType if this is updated

export function isTrialPlan(planName: string) {
  return planName === "free_trial";
}

export function isTeamPlan(planName: string) {
  return planName === "team";
}

export function isManagedPlan(planName: string) {
  return planName === "managed";
}

export function isFreePlan(planName: string) {
  return planName === "free_plan";
}

export function isProPlan(planName: string) {
  return planName === "pro_plan";
}

export function isStarterPlan(planName: string) {
  return planName === "starter";
}

export function isGrowthPlan(planName: string) {
  return planName === "growth";
}

export function isEnterprisePlan(planName: string) {
  return (
    !isTrialPlan(planName) &&
    !isTeamPlan(planName) &&
    !isManagedPlan(planName) &&
    !isFreePlan(planName) &&
    !isProPlan(planName) &&
    !isStarterPlan(planName) &&
    !isGrowthPlan(planName)
  );
}

export const PaidPlanTypes = {
  [V1BillingPlanType.BILLING_PLAN_TYPE_PRO]: true,
  [V1BillingPlanType.BILLING_PLAN_TYPE_TEAM]: true,
  [V1BillingPlanType.BILLING_PLAN_TYPE_ENTERPRISE]: true,
  [V1BillingPlanType.BILLING_PLAN_TYPE_STARTER]: true,
  [V1BillingPlanType.BILLING_PLAN_TYPE_GROWTH]: true,
};

export function getSubscriptionResumedText(endDate: string) {
  const date = DateTime.fromJSDate(new Date(endDate));
  if (!date.isValid || date.toMillis() < Date.now()) {
    return m.billing_today();
  }
  const resumeDate = date.plus({ day: 1 });
  return m.billing_on_date({ date: resumeDate.toLocaleString(DateTime.DATE_MED) });
}

// Since this could be triggered in a route that could be navigated from,
// we add a global and show it in org route's layout
export const showWelcomeToRillDialog = writable(false);
export const showWelcomeToRillDialogForPlan = writable("");

export function triggerWelcomeToRillDialog(planName: string) {
  showWelcomeToRillDialog.set(true);
  showWelcomeToRillDialogForPlan.set(planName);
}

export function formatCredit(credits: number): string {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: 2,
  }).format(credits);
}
