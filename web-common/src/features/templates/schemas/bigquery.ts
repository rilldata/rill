import type { MultiStepFormSchema } from "./types";

export const bigquerySchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    google_application_credentials: {
      type: "string",
      title: "GCP Credentials",
      description: "Upload a JSON key file for a service account with BigQuery access",
      format: "file",
      "x-display": "file",
      "x-accept": ".json",
      "x-step": "connector",
    },
    project_id: {
      type: "string",
      title: "Project ID",
      description: "Google Cloud project ID (optional if specified in credentials)",
      "x-placeholder": "my-project-id",
      "x-step": "connector",
    },
    log_queries: {
      type: "boolean",
      title: "Log Queries",
      description: "Enable logging of all SQL queries (useful for debugging)",
      default: false,
      "x-step": "connector",
      "x-advanced": true,
    },
    sql: {
      type: "string",
      title: "SQL Query",
      description: "SQL query to extract data from BigQuery",
      "x-placeholder": "SELECT * FROM `project.dataset.table`;",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model Name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
      "x-placeholder": "my_model",
      "x-step": "source",
    },
  },
  required: ["google_application_credentials", "sql", "name"],
};
