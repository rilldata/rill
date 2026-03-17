import type { MultiStepFormSchema } from "./types";

export const openaiSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "OpenAI",
  "x-category": "ai",
  properties: {
    api_key: {
      type: "string",
      title: "API Key",
      description: "API key for connecting to OpenAI",
      "x-placeholder": "sk-...",
      "x-secret": true,
    },
    model: {
      type: "string",
      title: "Model",
      description: "The OpenAI model to use",
      "x-placeholder": "gpt-4o",
    },
  },
  required: ["api_key"],
};
