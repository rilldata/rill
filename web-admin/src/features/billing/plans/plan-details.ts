import type { PlanTier } from "@rilldata/web-admin/features/billing/plans/types.ts";
import { formatMemorySize } from "@rilldata/web-common/lib/number-formatting/memory-size.ts";
import { formatCompactInteger } from "@rilldata/web-common/lib/formatters.ts";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
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
  translate: (value: string) => string;
  formatter?: (value: string) => string;
};
type PlanHighlightQuotaKey = keyof V1Quotas | keyof V1OrganizationQuotas;
type PlanHighlightQuotas = Partial<
  Record<PlanHighlightQuotaKey, string | number>
>;

const PlansQuotas: Record<string, PlanQuota> = {
  apiCallsPerSeat: {
    translate: (value) => m.billing_quota_api_calls({ value }),
    formatter: (value) => formatCompactInteger(Number(value)),
  },
  projects: {
    translate: (value) => m.billing_quota_projects({ value }),
  },
  seats: {
    translate: (value) => m.billing_quota_seats({ value }),
  },
  slotsTotal: {
    translate: (value) => m.billing_quota_compute_units({ value }),
  },
  storageLimitBytesPerDeployment: {
    translate: (value) => m.billing_quota_managed_db({ value }),
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
      "highlight:1m_ai_tokens",
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
      "highlight:embedded_analytics",
      "highlight:bring_own_ai",
      "highlight:2m_ai_tokens",
      "quota:storageLimitBytesPerDeployment",
      "quota:slotsTotal",
    ],
  },
];

const HighlightTranslations: Record<string, () => string> = {
  "1m_ai_tokens": () => m.billing_highlight_1m_ai_tokens(),
  "2m_ai_tokens": () => m.billing_highlight_2m_ai_tokens(),
  embedded_analytics: () => m.billing_highlight_embedded_analytics(),
  bring_own_ai: () => m.billing_highlight_bring_own_ai(),
};

export function resolvePlanHighlights(
  plan: SelfServePlan,
  quotas: PlanHighlightQuotas,
) {
  return plan.highlights
    .map((h) => {
      if (h.startsWith("highlight:")) {
        const key = h.slice("highlight:".length);
        return HighlightTranslations[key]?.() ?? "";
      }
      if (h.startsWith("quota:")) {
        const quotaKey = h.slice("quota:".length);
        const planQuota = PlansQuotas[quotaKey];
        if (!planQuota) return "";
        const quota = quotas[quotaKey as PlanHighlightQuotaKey];
        if (quota == null) return "";
        const quotaValue = String(quota);
        const value = planQuota.formatter
          ? planQuota.formatter(quotaValue)
          : quotaValue;
        return planQuota.translate(value);
      }
      return "";
    })
    .filter(Boolean);
}

export function getTranslatedPlanDisplayName(name: string): string {
  switch (name) {
    case "starter":
      return m.billing_plan_name_starter();
    case "growth":
      return m.billing_plan_name_growth();
    default:
      return name;
  }
}

export function getTranslatedPlanTagline(name: string): string {
  switch (name) {
    case "starter":
      return m.billing_plan_tagline_starter();
    case "growth":
      return m.billing_plan_tagline_growth();
    default:
      return "";
  }
}

export function getTranslatedPlanPriceUnit(): string {
  return m.billing_plan_price_unit();
}

export const SELF_SERVE_PLANS_BY_NAME = Object.fromEntries(
  SELF_SERVE_PLANS.map((plan) => [plan.name, plan]),
);
