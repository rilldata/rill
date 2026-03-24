import type { MultiStepFormSchema } from "./types";

export const geminiSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Gemini",
  "x-category": "ai",
  properties: {
    api_key: {
      type: "string",
      title: "API Key",
      description: "API key for connecting to Gemini",
      "x-placeholder": "AIza...",
      "x-secret": true,
    },
    model: {
      type: "string",
      title: "Model",
      description: "The Gemini model to use",
      "x-placeholder": "gemini-2.5-flash",
    },
  },
  required: ["api_key"],
};
