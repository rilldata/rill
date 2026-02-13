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
    headers: {
      title: "Headers",
      description: "HTTP headers to include in the request",
      "x-display": "key-value",
      "x-placeholder": "Header name",
      "x-hint": "e.g. Authorization: Bearer &lt;token&gt;",
      "x-step": "connector",
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
