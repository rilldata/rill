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

export function isEnterprisePlan(planName: string) {
  return (
    !isTrialPlan(planName) && !isTeamPlan(planName) && !isManagedPlan(planName)
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
