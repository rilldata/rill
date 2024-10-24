import type { V1BillingPlan } from "@rilldata/web-admin/client";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
import { DateTime } from "luxon";
import { writable } from "svelte/store";

export function formatDataSizeQuota(
  storageLimitBytesPerDeployment: string,
): string {
  if (
    Number.isNaN(Number(storageLimitBytesPerDeployment)) ||
    storageLimitBytesPerDeployment === "-1"
  )
    return "";
  return `Max ${formatMemorySize(Number(storageLimitBytesPerDeployment))} / Project`;
}

export function isTrialPlan(plan: V1BillingPlan) {
  return plan.default && plan.trialPeriodDays;
}

export function isTeamPlan(plan: V1BillingPlan) {
  return plan.name === "Teams";
}

export function isPOCPlan(plan: V1BillingPlan) {
  return plan.name === "Custom";
}

export function getSubscriptionResumedText(endDate: string) {
  const date = DateTime.fromJSDate(new Date(endDate));
  if (!date.isValid || date.toMillis() < Date.now()) {
    return "today";
  }
  const resumeDate = date.plus({ day: 1 });
  return "on " + resumeDate.toLocaleString(DateTime.DATE_MED);
}

export function getPlanDisplayName(plan: V1BillingPlan) {
  if (isTrialPlan(plan)) {
    return "Trial Plan";
  }
  if (isTeamPlan(plan)) {
    return "Team Plan";
  }
  if (isPOCPlan(plan)) {
    return "POC Plan";
  }
  return "Enterprise Plan";
}

// Since this could be triggered in a route that could be navigated from,
// we add a global and show it in org route's layout
export const showWelcomeToRillDialog = writable(false);
