import type { MultiStepFormSchema } from "./types";

export const publicSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Public URL",
  "x-category": "fileStore",
  "x-driver": "https",
  properties: {
    path: {
      type: "string",
      title: "Path",
      description: "Public HTTP/HTTPS URL to the remote file",
      pattern: "^https?://.+",
      errorMessage: {
        pattern: "Path must start with http:// or https://",
      },
      "x-placeholder": "https://example.com/data.csv",
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
