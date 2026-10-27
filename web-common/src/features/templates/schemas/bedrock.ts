import type { MultiStepFormSchema } from "./types";

export const bedrockSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Amazon Bedrock",
  "x-category": "ai",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      description: "Choose how to authenticate to Amazon Bedrock",
      enum: ["api_key", "access_keys"],
      default: "api_key",
      "x-display": "radio",
      "x-enum-labels": ["API key", "AWS access keys"],
      "x-ui-only": true,
    },
    api_key: {
      type: "string",
      title: "API key",
      description:
        "Long-term Amazon Bedrock API key. It only works in the region it was generated in.",
      "x-placeholder": "Enter your Bedrock API key",
      "x-secret": true,
      "x-visible-if": { auth_method: "api_key" },
    },
    aws_access_key_id: {
      type: "string",
      title: "AWS access key ID",
      description: "AWS access key ID with access to Bedrock",
      "x-placeholder": "your_access_key_id",
      "x-secret": true,
      "x-visible-if": { auth_method: "access_keys" },
    },
    aws_secret_access_key: {
      type: "string",
      title: "AWS secret access key",
      description: "AWS secret access key paired with the access key ID",
      "x-placeholder": "your_secret_access_key",
      "x-secret": true,
      "x-visible-if": { auth_method: "access_keys" },
    },
    model: {
      type: "string",
      title: "Model",
      description: "The Bedrock model ID, inference profile ID, or ARN to use",
      "x-placeholder": "us.anthropic.claude-sonnet-4-5-20250929-v1:0",
    },
    region: {
      type: "string",
      title: "AWS Region",
      description: "AWS region to call Bedrock in",
      "x-placeholder": "us-east-1",
    },
  },
  required: ["model"],
  allOf: [
    {
      if: { properties: { auth_method: { const: "api_key" } } },
      then: { required: ["api_key"] },
    },
    {
      if: { properties: { auth_method: { const: "access_keys" } } },
      then: { required: ["aws_access_key_id", "aws_secret_access_key"] },
    },
  ],
};
