import type { MultiStepFormSchema } from "./types";

export const bigquerySchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      enum: ["service_account"],
      default: "service_account",
      description: "Choose how to authenticate to BigQuery",
      "x-display": "radio",
      "x-enum-labels": ["Service Account"],
      "x-enum-descriptions": [
        "Upload a JSON key file for a service account with BigQuery access.",
      ],
      "x-grouped-fields": {
        service_account: ["google_application_credentials", "project_id"],
      },
    },
    google_application_credentials: {
      type: "string",
      title: "Service Account Key",
      description: "Upload a JSON key file for a service account with BigQuery access",
      format: "file",
      "x-display": "file",
      "x-accept": ".json",
      "x-visible-if": { auth_method: "service_account" },
    },
    project_id: {
      type: "string",
      title: "Project ID",
      description: "Google Cloud project ID (optional if specified in credentials)",
      "x-placeholder": "my-project-id",
      "x-visible-if": { auth_method: "service_account" },
    },
  },
  allOf: [
    {
      if: { properties: { auth_method: { const: "service_account" } } },
      then: { required: ["google_application_credentials"] },
    },
  ],
};
