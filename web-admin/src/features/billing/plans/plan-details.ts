import type { PlanTier } from "@rilldata/web-admin/features/billing/plans/types.ts";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size.ts";
import { formatCompactInteger } from "@rilldata/web-common/lib/formatters.ts";
import type {
  V1OrganizationQuotas,
  V1Quotas,
} from "@rilldata/web-admin/client";

export type SelfServePlan = {
  tier: Extract<PlanTier, "starter" | "growth">;
  // name must match the plan name in the billing system (see admin/billing/orb.go::getPlanType).
  name: string;
  displayName: string;
  // Price and price unit are displayed in the UI slightly differently.
  // Otherwise, it would be redundant to have different keys.
  price: string;
  priceUnit: string;
  tagline: string;
  highlights: string[];
  recommended?: boolean;
};

type PlanQuota = {
  name: string;
  template: string;
  formatter?: (value: string) => string;
};
type PlanHighlightQuotaKey = keyof V1Quotas | keyof V1OrganizationQuotas;
type PlanHighlightQuotas = Partial<
  Record<PlanHighlightQuotaKey, string | number>
>;

const PlansQuotas: Record<string, PlanQuota> = {
  apiCallsPerSeat: {
    name: "API calls",
    template: "{value} API calls / seat / month",
    formatter: (value) => formatCompactInteger(Number(value)),
  },
  projects: {
    name: "Projects",
    template: "Up to {value} projects",
  },
  seats: {
    name: "Seats",
    template: "Up to {value} seats",
  },
  slotsTotal: {
    name: "Compute units",
    template: "Up to {value} compute units",
  },
  storageLimitBytesPerDeployment: {
    name: "Managed database size",
    template: "Managed database up to {value}",
    formatter: (value) => formatMemorySize(Number(value)),
  },
};

// Marketing highlights shown in the upgrade chooser. The pricing and quota copy is
// intentionally summarised here; the authoritative plan limits come from the billing system.
export const SELF_SERVE_PLANS: SelfServePlan[] = [
  {
    tier: "starter",
    name: "starter",
    displayName: "Starter",
    price: "$20",
    priceUnit: "/ seat / month",
    tagline: "For small teams getting started.",
    highlights: [
      "quota:seats",
      "quota:projects",
      "1M AI tokens / seat / month",
      "quota:apiCallsPerSeat",
      "quota:storageLimitBytesPerDeployment",
      "quota:slotsTotal",
    ],
  },
  {
    tier: "growth",
    name: "growth",
    displayName: "Growth",
    price: "$30",
    priceUnit: "/ seat / month",
    tagline: "For growing teams and embedded analytics.",
    recommended: true,
    highlights: [
      "quota:seats",
      "quota:projects",
      "Embedded analytics",
      "Bring your own AI model",
      "2M AI tokens / seat / month",
      "quota:storageLimitBytesPerDeployment",
      "quota:slotsTotal",
    ],
  },
];

const ValueRegex = /{value}/g;
const QuotaPrefix = "quota:";
const QuotaLength = QuotaPrefix.length;
export function resolvePlanHighlights(
  plan: SelfServePlan,
  quotas: PlanHighlightQuotas,
) {
  return plan.highlights
    .map((h) => {
      if (!h.startsWith(QuotaPrefix)) return h;
      const quotaKey = h.slice(QuotaLength);
      const planQuota = PlansQuotas[quotaKey];
      if (!planQuota) return "";
      const quota = quotas[quotaKey as PlanHighlightQuotaKey];
      if (quota == null) return "";
      const quotaValue = String(quota);
      const value = planQuota.formatter
        ? planQuota.formatter(quotaValue)
        : quotaValue;
      return planQuota.template.replace(ValueRegex, value);
    })
    .filter(Boolean);
}

export const SELF_SERVE_PLANS_BY_NAME = Object.fromEntries(
  SELF_SERVE_PLANS.map((plan) => [plan.name, plan]),
);
