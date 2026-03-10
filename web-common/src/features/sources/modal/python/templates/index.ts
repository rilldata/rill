import { script as ga4Script } from "./ga4";
import { script as stripeScript } from "./stripe";
import { script as orbScript } from "./orb";
import { script as httpScript } from "./http";
import { script as blankScript } from "./blank";

export interface PythonTemplate {
  id: string;
  label: string;
  description: string;
  defaultPath: string;
  suggestedSecrets: string[];
  script: string;
}

export const pythonTemplates: PythonTemplate[] = [
  {
    id: "ga4",
    label: "Google Analytics (GA4)",
    description: "Sessions, users, page views by date and channel",
    defaultPath: "scripts/google_analytics.py",
    suggestedSecrets: ["gcs"],
    script: ga4Script,
  },
  {
    id: "stripe",
    label: "Stripe",
    description: "Charges, customers, and subscription data",
    defaultPath: "scripts/stripe_charges.py",
    suggestedSecrets: [],
    script: stripeScript,
  },
  {
    id: "orb",
    label: "Orb",
    description: "Usage events and billing data from Orb",
    defaultPath: "scripts/orb_usage.py",
    suggestedSecrets: [],
    script: orbScript,
  },
  {
    id: "http",
    label: "REST API",
    description: "Generic HTTP endpoint data extraction",
    defaultPath: "scripts/http_api.py",
    suggestedSecrets: [],
    script: httpScript,
  },
  {
    id: "blank",
    label: "Blank Script",
    description: "Minimal template with the Rill output contract",
    defaultPath: "scripts/extract.py",
    suggestedSecrets: [],
    script: blankScript,
  },
];
