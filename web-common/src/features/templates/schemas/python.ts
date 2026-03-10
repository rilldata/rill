import type { MultiStepFormSchema } from "./types";

export const pythonSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Python",
  "x-category": "fileStore",
  "x-form-width": "wide",
  properties: {
    source_mode: {
      type: "string",
      title: "Script source",
      enum: ["template", "custom"],
      default: "template",
      "x-display": "tabs",
      "x-enum-labels": ["Template", "Custom script"],
      "x-ui-only": true,
      "x-tab-group": {
        template: ["template_id"],
        custom: ["code_path"],
      },
      "x-step": "source",
    },
    template_id: {
      type: "string",
      title: "Template",
      description: "Select a starter template",
      enum: ["ga4", "stripe", "orb", "http", "blank"],
      "x-display": "select",
      "x-select-style": "rich",
      "x-enum-labels": [
        "Google Analytics (GA4)",
        "Stripe",
        "Orb",
        "REST API",
        "Blank Script",
      ],
      "x-enum-descriptions": [
        "Sessions, users, page views by date and channel",
        "Charges, customers, and subscription data",
        "Usage events and billing data from Orb",
        "Generic HTTP endpoint data extraction",
        "Minimal template with the Rill output contract",
      ],
      "x-ui-only": true,
      "x-visible-if": { source_mode: "template" },
      "x-step": "source",
    },
    code_path: {
      type: "string",
      title: "Script path",
      description:
        "Path to an existing Python script relative to the project root.",
      "x-placeholder": "scripts/extract.py",
      "x-monospace": true,
      "x-visible-if": { source_mode: "custom" },
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name of the model to create",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_python_model",
      "x-step": "source",
    },
    create_secrets_from_connectors: {
      type: "string",
      title: "Pass secrets from connectors",
      description:
        "Inject credentials from existing connectors as environment variables in your script.",
      "x-placeholder": "Select a connector...",
      "x-hint":
        "Connector credentials are mapped to standard env vars (e.g. GCS → GOOGLE_APPLICATION_CREDENTIALS).",
      "x-step": "source",
    },
  },
  required: ["name"],
  allOf: [
    {
      if: { properties: { source_mode: { const: "template" } } },
      then: { required: ["template_id"] },
    },
    {
      if: { properties: { source_mode: { const: "custom" } } },
      then: { required: ["code_path"] },
    },
  ],
};
