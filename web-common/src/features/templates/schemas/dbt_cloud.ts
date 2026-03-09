import type { MultiStepFormSchema } from "./types";

export const dbtCloudSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "dbt Cloud",
  "x-category": "dbt",
  properties: {
    api_token: {
      type: "string",
      title: "API Token",
      description: "dbt Cloud API token for authentication",
      "x-placeholder": "abc123def456ghi789",
      "x-secret": true,
      "x-env-var-name": "DBT_CLOUD_API_TOKEN",
    },
    account_id: {
      type: "string",
      title: "Account ID",
      description: "Your dbt Cloud account ID",
      "x-placeholder": "70471823540700",
    },
    environment_id: {
      type: "string",
      title: "Environment ID",
      description: "The dbt Cloud environment to fetch manifests from",
      "x-placeholder": "70471823525999",
    },
    base_url: {
      type: "string",
      title: "Base URL",
      description: "dbt Cloud host URL; defaults to https://cloud.getdbt.com",
      "x-placeholder": "https://cloud.getdbt.com",
    },
    webhook_secret: {
      type: "string",
      title: "Webhook Secret",
      description:
        "HMAC secret for validating dbt Cloud webhook payloads (optional)",
      "x-secret": true,
      "x-placeholder": "supersecretwebhookkey",
      "x-env-var-name": "DBT_CLOUD_WEBHOOK_SECRET",
    },
  },
  required: ["api_token", "account_id", "environment_id"],
};
