import type { PlanTier } from "@rilldata/web-admin/features/billing/plans/types.ts";

export type SelfServePlan = {
  tier: Extract<PlanTier, "starter" | "growth">;
  // name must match the plan name in the billing system (see admin/billing/orb.go::getPlanType).
  name: string;
  displayName: string;
  price: string;
  priceUnit: string;
  tagline: string;
  highlights: string[];
  recommended?: boolean;
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
      "Up to 20 seats",
      "Up to 3 projects",
      "1M AI tokens / seat / month",
      "2,500 API calls / seat / month",
      "Managed database up to 10 GB",
      "Up to 32 compute units",
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
      "Up to 100 seats",
      "Up to 10 projects",
      "Embedded analytics",
      "Bring your own AI model",
      "2M AI tokens / seat / month",
      "Managed database up to 1 TB",
      "Up to 128 compute units",
    ],
  },
];
