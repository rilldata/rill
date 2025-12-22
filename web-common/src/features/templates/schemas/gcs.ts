import type { MultiStepFormSchema } from "./types";

export const gcsSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      enum: ["credentials", "hmac", "public"],
      default: "credentials",
      description: "Choose how to authenticate to GCS",
      "x-display": "radio",
      "x-enum-labels": ["GCP credentials", "HMAC keys", "Public"],
      "x-enum-descriptions": [
        "Upload a JSON key file for a service account with GCS access.",
        "Use HMAC access key and secret for S3-compatible authentication.",
        "Access publicly readable buckets without credentials.",
      ],
      "x-grouped-fields": {
        credentials: ["google_application_credentials"],
        hmac: ["key_id", "secret"],
        public: [],
      },
      "x-step": "connector",
    },
    google_application_credentials: {
      type: "string",
      title: "Service account key",
      description:
        "Upload a JSON key file for a service account with GCS access.",
      format: "file",
      "x-display": "file",
      "x-accept": ".json",
      "x-step": "connector",
      "x-visible-if": { auth_method: "credentials" },
    },
    key_id: {
      type: "string",
      title: "Access Key ID",
      description: "HMAC access key ID for S3-compatible authentication",
      "x-placeholder": "Enter your HMAC access key ID",
      "x-secret": true,
      "x-step": "connector",
      "x-visible-if": { auth_method: "hmac" },
    },
    secret: {
      type: "string",
      title: "Secret Access Key",
      description: "HMAC secret access key for S3-compatible authentication",
      "x-placeholder": "Enter your HMAC secret access key",
      "x-secret": true,
      "x-step": "connector",
      "x-visible-if": { auth_method: "hmac" },
    },
    path_prefixes: {
      type: "string",
      title: "Path prefixes",
      description:
        "List of prefixes to filter the files in the GCS bucket. Leave empty to include all files.",
      "x-step": "connector",
      "x-advanced": true,
    },
    allow_host_access: {
      type: "boolean",
      title: "Allow host access",
      description:
        "Allow access to the GCS bucket from the host machine. This is useful for debugging and testing.",
      default: false,
      "x-step": "connector",
      "x-advanced": true,
    },
    path: {
      type: "string",
      title: "GCS URI",
      description: "Path to your GCS bucket or prefix",
      pattern: "^gs://",
      "x-placeholder": "gs://bucket/path",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
      "x-placeholder": "my_model",
      "x-step": "source",
    },
  },
  required: ["path", "name"],
  allOf: [
    {
      if: { properties: { auth_method: { const: "credentials" } } },
      then: { required: ["google_application_credentials", "path", "name"] },
    },
    {
      if: { properties: { auth_method: { const: "hmac" } } },
      then: { required: ["key_id", "secret", "path", "name"] },
    },
    {
      if: { properties: { auth_method: { const: "public" } } },
      then: { required: ["path", "name"] },
    },
  ],
};
