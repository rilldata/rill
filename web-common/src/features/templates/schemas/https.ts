import type { MultiStepFormSchema } from "./types";

export const httpsSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "HTTP(S)",
  "x-category": "fileStore",
  properties: {
    headers: {
      type: "string",
      title: "Headers",
      description:
        "HTTP headers to include in requests (one per line, format: Header-Name: value)",
      "x-display": "textarea",
      "x-placeholder": "Authorization: Bearer <token>",
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
  required: ["headers", "path", "name"],
};
