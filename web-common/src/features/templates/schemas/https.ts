import type { MultiStepFormSchema } from "./types";

export const httpsSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      enum: ["public", "headers"],
      default: "public",
      description: "Choose how to authenticate to the REST API",
      "x-display": "radio",
      "x-enum-labels": ["Public", "Custom Headers"],
      "x-enum-descriptions": [
        "Access publicly available APIs without authentication.",
        "Provide custom HTTP headers for authentication (e.g., Authorization, API keys).",
      ],
      "x-grouped-fields": {
        public: [],
        headers: ["headers"],
      },
    },
    headers: {
      type: "string",
      title: "HTTP Headers (JSON)",
      description:
        'HTTP headers as JSON object. Example: {"Authorization": "Bearer my-token", "X-API-Key": "value"}',
      "x-placeholder": '{"Authorization": "Bearer my-token"}',
      "x-display": "textarea",
      "x-visible-if": { auth_method: "headers" },
    },
  },
  allOf: [
    {
      if: { properties: { auth_method: { const: "headers" } } },
      then: { required: ["headers"] },
    },
  ],
};
