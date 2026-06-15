type Plan = {
  name: string;
  title: string;
  main: string;
  sub: string;
  paid?: boolean;
  custom?: boolean;
  features: string[];
};

export const LegacyTrialPlan: Plan = {
  name: "free_trial",
  title: "Trial",
  main: "30 day free trial",
  sub: "No credit card required",
  features: [
    "30-day free trial period",
    "Self-serve compute units (2 units minimum)",
    "1 GB storage included · $1/GB above",
    `"Made with Rill" badge`,
  ],
};
export const CreditsTrialPlan: Plan = {
  name: "free_plan",
  title: "Free",
  main: "$250 free credit",
  sub: "No time limit",
  features: [
    "Credit rolls over when you subscribe to Pro",
    "Self-serve compute units (2 units minimum)",
    "1 GB storage included · $1/GB above",
    `"Made with Rill" badge`,
  ],
};

export const LegacyTeamPlan: Plan = {
  name: "team",
  title: "Team plan (legacy)",
  main: "$250/mo",
  sub: "Flat rate + storage",
  paid: true,
  features: [
    "$250/mo flat charge",
    "10 GB storage included · $25/GB over",
    "Email support",
    `"Made with Rill" badge`,
  ],
};

export const StarterPaidPlan: Plan = {
  name: "starter",
  title: "Starter",
  main: "$20 / user / month",
  sub: "For teams just getting started",
  paid: true,
  features: [
    "$20 / user / month, Up to 50 users",
    "Up to 3 projects",
    "$0.15 / compute-unit per hour (1 unit minimum)",
  ],
};

export const GrowthPaidPlan: Plan = {
  name: "growth",
  title: "Growth",
  main: "$30 / user / month ⭐ Most Popular",
  sub: "For scaling teams",
  paid: true,
  features: [
    "$30 / user / month, Up to 200 users",
    "Up to 20 projects",
    "$0.20 / compute-unit per hour (1 unit minimum)",
    `Custom branding with "Powered by Rill"`,
  ],
};

export const EnterprisePlan: Plan = {
  name: "enterprise",
  title: "Enterprise",
  main: "Custom Pricing",
  sub: "Negotiated contract",
  custom: true,
  features: [
    "Negotiated seats",
    "Unlimited projects",
    `Custom branding with custom logo`,
  ],
};
