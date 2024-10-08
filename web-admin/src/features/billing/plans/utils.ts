import type { V1BillingPlan } from "@rilldata/web-admin/client";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";
import { DateTime } from "luxon";

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

export function getSubscriptionResumedText(endDate: string) {
  const date = DateTime.fromJSDate(new Date(endDate));
  if (!date.isValid || date.toMillis() < Date.now()) {
    return "today";
  }
  const resumeDate = date.plus({ day: 1 });
  return "on " + resumeDate.toLocaleString(DateTime.DATE_MED);
}
