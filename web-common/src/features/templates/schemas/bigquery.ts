import type { MultiStepFormSchema } from "./types";

export const bigquerySchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    google_application_credentials: {
      type: "string",
      title: "Service Account Key",
      description: "Upload a JSON key file for a service account with BigQuery access",
      format: "file",
      "x-display": "file",
      "x-accept": ".json",
    },
    project_id: {
      type: "string",
      title: "Project ID",
      description: "Google Cloud project ID (optional if specified in credentials)",
      "x-placeholder": "my-project-id",
    },
  },
  required: ["google_application_credentials"],
};
