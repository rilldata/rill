import type { MultiStepFormSchema } from "./types";

export const claudeSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Claude",
  "x-category": "ai",
  properties: {
    api_key: {
      type: "string",
      title: "API Key",
      description: "API key for connecting to Claude",
      "x-placeholder": "sk-ant-...",
      "x-secret": true,
    },
    model: {
      type: "string",
      title: "Model",
      description: "The Claude model to use",
      "x-placeholder": "claude-sonnet-4-5-20250929",
    },
  },
  required: ["api_key"],
};
