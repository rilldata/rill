import type { MultiStepFormSchema } from "./types";

export const s3Schema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      description: "Choose how to authenticate to S3",
      enum: ["access_keys", "public"],
      default: "access_keys",
      "x-display": "radio",
      "x-enum-labels": ["Access keys", "Public"],
      "x-enum-descriptions": [
        "Use AWS access key ID and secret access key.",
        "Access publicly readable buckets without credentials.",
      ],
      "x-step": "connector",
    },
    aws_access_key_id: {
      type: "string",
      title: "Access Key ID",
      description: "AWS access key ID for the bucket",
      "x-placeholder": "Enter AWS access key ID",
      "x-secret": true,
      "x-step": "connector",
      "x-visible-if": { auth_method: "access_keys" },
    },
    aws_secret_access_key: {
      type: "string",
      title: "Secret Access Key",
      description: "AWS secret access key for the bucket",
      "x-placeholder": "Enter AWS secret access key",
      "x-secret": true,
      "x-step": "connector",
      "x-visible-if": { auth_method: "access_keys" },
    },
    region: {
      type: "string",
      title: "Region",
      description:
        "Rill uses your default AWS region unless you set it explicitly.",
      "x-placeholder": "us-east-1",
      "x-step": "connector",
      "x-visible-if": { auth_method: "access_keys" },
    },
    endpoint: {
      type: "string",
      title: "Endpoint",
      description:
        "Override the S3 endpoint (for S3-compatible services like R2/MinIO).",
      "x-placeholder": "https://s3.example.com",
      "x-step": "connector",
      "x-visible-if": { auth_method: "access_keys" },
    },
    aws_role_arn: {
      type: "string",
      title: "AWS Role ARN",
      description: "AWS Role ARN to assume",
      "x-placeholder": "arn:aws:iam::123456789012:role/MyRole",
      "x-secret": true,
      "x-step": "connector",
      "x-visible-if": { auth_method: "access_keys" },
    },
    path: {
      type: "string",
      title: "S3 URI",
      description: "Path to your S3 bucket or prefix",
      pattern: "^s3://",
      "x-placeholder": "s3://bucket/path",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
      "x-placeholder": "my_model",
      "x-step": "source",
    },
  },
  allOf: [
    {
      if: { properties: { auth_method: { const: "access_keys" } } },
      then: {
        required: ["aws_access_key_id", "aws_secret_access_key"],
      },
    },
  ],
};
