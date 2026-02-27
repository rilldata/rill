import GoogleBigQuery from "../../../components/icons/connectors/GoogleBigQuery.svelte";
import GoogleBigQueryIcon from "../../../components/icons/connectors/GoogleBigQueryIcon.svelte";
import type { MultiStepFormSchema } from "./types";

export const bigquerySchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "BigQuery",
  "x-category": "warehouse",
  "x-icon": GoogleBigQuery,
  "x-small-icon": GoogleBigQueryIcon,
  properties: {
    google_application_credentials: {
      type: "string",
      title: "GCP credentials",
      description: "Service account JSON (uploaded or pasted)",
      format: "file",
      "x-display": "file",
      "x-file-accept": ".json",
      "x-file-encoding": "json",
      "x-file-extract": { project_id: "project_id" },
      "x-secret": true,
      "x-env-var-name": "GOOGLE_APPLICATION_CREDENTIALS",
      "x-step": "connector",
    },
    project_id: {
      type: "string",
      title: "Project ID",
      description: "Google Cloud project ID to use for queries",
      "x-placeholder": "my-project",
      "x-hint":
        "If empty, Rill will use the project ID from your credentials when available.",
      "x-step": "connector",
    },
    sql: {
      type: "string",
      title: "SQL",
      description: "SQL query to run against your warehouse",
      "x-placeholder": "Input SQL",
      "x-step": "explorer",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_model",
      "x-step": "explorer",
    },
  },
  required: ["google_application_credentials", "project_id", "sql", "name"],
};
