import type { V1BillingPlan } from "@rilldata/web-admin/client";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size";

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
