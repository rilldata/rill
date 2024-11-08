import type { V1BillingPlan } from "@rilldata/web-admin/client";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
import { DateTime } from "luxon";
import { writable } from "svelte/store";

export function formatDataSizeQuota(
  totalSizeInBytes: number,
  storageLimitBytesPerDeployment: string,
): string {
  const maxSize = Number(storageLimitBytesPerDeployment);
  if (Number.isNaN(maxSize) || storageLimitBytesPerDeployment === "-1")
    return "";
  const formattedTotal = formatMemorySize(totalSizeInBytes);
  const formattedMax = formatMemorySize(maxSize);
  const percent =
    formattedTotal > formattedMax
      ? "100+"
      : Math.round((totalSizeInBytes * 100) / maxSize) + "";
  return `${formattedTotal} of ${formattedMax} (${percent}%)`;
}

export function isTrialPlan(plan: V1BillingPlan) {
  return plan.name === "free_trial";
}

export function isTeamPlan(plan: V1BillingPlan) {
  return plan.name === "team";
}

export function isPOCPlan(plan: V1BillingPlan) {
  return plan.name === "poc";
}

export function isEnterprisePlan(plan: V1BillingPlan) {
  return !isTrialPlan(plan) && !isTeamPlan(plan) && !isPOCPlan(plan);
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
