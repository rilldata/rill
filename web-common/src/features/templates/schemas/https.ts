import type { MultiStepFormSchema } from "./types";

export const httpsSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "HTTP(S)",
  "x-category": "fileStore",
  "x-button-labels": {
    "*": { "*": { idle: "Continue", loading: "Continuing..." } },
  },
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      enum: ["with_headers", "public"],
      default: "with_headers",
      description: "Choose how to authenticate to GCS",
      "x-display": "radio",
      "x-enum-labels": ["Headers", "Public"],
      "x-enum-descriptions": [
        "Add headers for credentials.",
        "Access publicly readable urls without credentials.",
      ],
      "x-ui-only": true,
      "x-grouped-fields": {
        with_headers: ["headers"],
        public: [],
      },
      "x-step": "connector",
    },
    headers: {
      title: "Headers",
      description: "HTTP headers to include in the request",
      "x-display": "key-value",
      "x-placeholder": "Header name",
      "x-hint": "e.g. Authorization: Bearer &lt;token&gt;",
      "x-step": "connector",
      "x-visible-if": { auth_method: "with_headers" },
    },
    path: {
      type: "string",
      title: "URI",
      description: "HTTP/HTTPS URL to the remote file",
      pattern: "^https?://.+",
      errorMessage: {
        pattern: "URI must start with http:// or https://",
      },
      "x-placeholder": "https://example.com/file.csv",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name of the model",
      "x-placeholder": "my_new_model",
      "x-step": "source",
    },
  },
  required: ["path", "name"],
};
