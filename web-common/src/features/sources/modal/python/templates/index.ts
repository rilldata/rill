import { script as ga4Script } from "./ga4";
import { script as stripeScript } from "./stripe";
import { script as orbScript } from "./orb";
import { script as httpScript } from "./http";
import { script as hubspotScript } from "./hubspot";
import { script as blankScript } from "./blank";

export interface TemplateEnvVar {
  key: string;
  label: string;
  placeholder: string;
}

export interface PythonTemplate {
  id: string;
  label: string;
  description: string;
  defaultPath: string;
  suggestedSecrets: string[];
  envVars: TemplateEnvVar[];
  script: string;
}

export const pythonTemplates: PythonTemplate[] = [
  {
    id: "ga4",
    label: "Google Analytics (GA4)",
    description: "Sessions, users, page views by date and channel",
    defaultPath: "scripts/google_analytics.py",
    suggestedSecrets: ["gcs"],
    envVars: [
      {
        key: "GA4_PROPERTY_ID",
        label: "GA4 Property ID",
        placeholder: "123456789",
      },
    ],
    script: ga4Script,
  },
  {
    id: "stripe",
    label: "Stripe",
    description: "Charges, customers, and subscription data",
    defaultPath: "scripts/stripe_charges.py",
    suggestedSecrets: [],
    envVars: [],
    script: stripeScript,
  },
  {
    id: "orb",
    label: "Orb",
    description: "Usage events and billing data from Orb",
    defaultPath: "scripts/orb_usage.py",
    suggestedSecrets: [],
    envVars: [],
    script: orbScript,
  },
  {
    id: "hubspot",
    label: "HubSpot",
    description: "Contacts, companies, and CRM data",
    defaultPath: "scripts/hubspot_contacts.py",
    suggestedSecrets: [],
    envVars: [
      {
        key: "HUBSPOT_ACCESS_TOKEN",
        label: "HubSpot Access Token",
        placeholder: "pat-na1-xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
      },
    ],
    script: hubspotScript,
  },
  {
    id: "http",
    label: "REST API",
    description: "Generic HTTP endpoint data extraction",
    defaultPath: "scripts/http_api.py",
    suggestedSecrets: [],
    envVars: [],
    script: httpScript,
  },
  {
    id: "blank",
    label: "Blank Script",
    description: "Minimal template with the Rill output contract",
    defaultPath: "scripts/extract.py",
    suggestedSecrets: [],
    envVars: [],
    script: blankScript,
  },
];
