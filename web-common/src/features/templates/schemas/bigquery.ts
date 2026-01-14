import type { MultiStepFormSchema } from "./types";

export const bigquerySchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    google_application_credentials: {
      type: "string",
      title: "GCP credentials",
      description: "Service account JSON (uploaded or pasted)",
      format: "file",
      "x-display": "file",
      "x-accept": ".json",
      "x-secret": true,
    },
    project_id: {
      type: "string",
      title: "Project ID",
      description: "Google Cloud project ID to use for queries",
      "x-placeholder": "my-project",
      "x-hint":
        "If empty, Rill will use the project ID from your credentials when available.",
    },
  },
  required: ["google_application_credentials", "project_id"],
};
