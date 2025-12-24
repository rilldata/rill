import type { MultiStepFormSchema } from "./types";

export const httpsSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      enum: ["headers", "public"],
      default: "headers",
      description: "Choose how to authenticate to the REST API",
      "x-display": "radio",
      "x-enum-labels": ["Custom Headers", "Public"],
      "x-enum-descriptions": [
        "Access publicly available APIs without authentication.",
        "Provide custom HTTP headers for authentication (e.g., Authorization, API keys).",
      ],
      "x-grouped-fields": {
        headers: ["headers"],
      },
      "x-step": "connector",
    },
    headers: {
      type: "string",
      title: "HTTP Headers (JSON)",
      description:
        'HTTP headers as JSON object. Example: {"Authorization": "Bearer my-token", "X-API-Key": "value"}',
      "x-placeholder": '{"Authorization": "Bearer my-token"}',
      "x-step": "connector",
      "x-visible-if": { auth_method: "headers" },
    },
    path: {
      type: "string",
      title: "URL",
      description: "HTTP(S) URL to fetch data from",
      pattern: "^https?://",
      "x-placeholder": "https://api.example.com/data",
      "x-step": "source",
    },
    format: {
      type: "string",
      title: "Data Format",
      description: "Format of the data returned by the API",
      enum: ["json", "csv"],
      default: "json",
      "x-display": "radio",
      "x-enum-labels": ["JSON", "CSV"],
      "x-enum-descriptions": ["JSON format", "CSV format"],
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
  required: ["path", "name"],
  allOf: [
    {
      if: { properties: { auth_method: { const: "headers" } } },
      then: { required: ["headers", "path", "name"] },
    },
    {
      if: { properties: { auth_method: { const: "public" } } },
      then: { required: ["path", "name"] },
    },
  ],
};
