import type { MultiStepFormSchema } from "./types";

export const s3Schema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "S3",
  "x-category": "objectStore",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      description: "Choose how to authenticate to S3",
      enum: ["access_keys", "gcp_web_identity", "web_identity_file", "public"],
      default: "access_keys",
      "x-display": "radio",
      "x-enum-labels": [
        "Access keys",
        "GCP Workload Identity",
        "OIDC token file",
        "Public",
      ],
      "x-enum-descriptions": [
        "Use AWS access key ID and secret access key.",
        "Exchange a Google-signed workload identity token for AWS credentials.",
        "Exchange an OIDC token from a mounted file for AWS credentials.",
        "Access publicly readable buckets without credentials.",
      ],
      "x-ui-only": true,
      "x-grouped-fields": {
        access_keys: [
          "aws_access_key_id",
          "aws_secret_access_key",
          "aws_access_token",
          "region",
          "endpoint",
          "aws_role_arn",
          "aws_role_session_name",
          "aws_external_id",
        ],
        gcp_web_identity: [
          "gcp_workload_identity_audience",
          "aws_web_identity_role_arn",
          "aws_web_identity_role_session_name",
          "aws_role_arn",
          "aws_role_session_name",
          "aws_external_id",
          "region",
        ],
        web_identity_file: [
          "aws_web_identity_token_file",
          "aws_web_identity_role_arn",
          "aws_web_identity_role_session_name",
          "aws_role_arn",
          "aws_role_session_name",
          "aws_external_id",
          "region",
        ],
        public: [],
      },
      "x-step": "connector",
    },
    aws_access_key_id: {
      type: "string",
      title: "Access Key ID",
      description: "AWS access key ID for the bucket",
      "x-placeholder": "Enter AWS access key ID",
      "x-secret": true,
      "x-env-var-name": "AWS_ACCESS_KEY_ID",
      "x-step": "connector",
      "x-visible-if": { auth_method: "access_keys" },
    },
    aws_secret_access_key: {
      type: "string",
      title: "Secret Access Key",
      description: "AWS secret access key for the bucket",
      "x-placeholder": "Enter AWS secret access key",
      "x-secret": true,
      "x-env-var-name": "AWS_SECRET_ACCESS_KEY",
      "x-step": "connector",
      "x-visible-if": { auth_method: "access_keys" },
    },
    aws_access_token: {
      type: "string",
      title: "Session Token",
      description:
        "Optional AWS session token when the access key is temporary",
      "x-placeholder": "Enter AWS session token",
      "x-secret": true,
      "x-env-var-name": "AWS_SESSION_TOKEN",
      "x-step": "connector",
      "x-visible-if": { auth_method: "access_keys" },
      "x-advanced": true,
    },
    region: {
      type: "string",
      title: "Region",
      description:
        "Rill uses your default AWS region unless you set it explicitly.",
      "x-placeholder": "us-east-1",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: ["access_keys", "gcp_web_identity", "web_identity_file"],
      },
    },
    endpoint: {
      type: "string",
      title: "Endpoint",
      description:
        "Override the S3 endpoint (for S3-compatible services like R2/MinIO).",
      "x-placeholder": "https://s3.example.com",
      "x-step": "connector",
      "x-visible-if": { auth_method: "access_keys" },
      "x-advanced": true,
    },
    aws_role_arn: {
      type: "string",
      title: "AWS Role ARN",
      description:
        "Optional target AWS role to assume. With WebIdentity, this is a second role assumed using the federated role credentials.",
      "x-placeholder": "arn:aws:iam::123456789012:role/MyRole",
      "x-secret": true,
      "x-env-var-name": "AWS_ROLE_ARN",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: ["access_keys", "gcp_web_identity", "web_identity_file"],
      },
      "x-advanced": true,
    },
    aws_role_session_name: {
      type: "string",
      title: "Role Session Name",
      description: "Session name for STS AssumeRole",
      "x-placeholder": "rill-session",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: ["access_keys", "gcp_web_identity", "web_identity_file"],
      },
      "x-advanced": true,
    },
    aws_external_id: {
      type: "string",
      title: "External ID",
      description: "External ID for cross-account role assumption",
      "x-placeholder": "your-external-id",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: ["access_keys", "gcp_web_identity", "web_identity_file"],
      },
      "x-advanced": true,
    },
    aws_web_identity_token_file: {
      type: "string",
      title: "OIDC token file",
      description: "Path to a mounted file containing an OIDC identity token",
      "x-placeholder": "/var/run/secrets/oidc/token",
      "x-env-var-name": "AWS_WEB_IDENTITY_TOKEN_FILE",
      "x-step": "connector",
      "x-visible-if": { auth_method: "web_identity_file" },
    },
    aws_web_identity_role_arn: {
      type: "string",
      title: "WebIdentity role ARN",
      description: "AWS role whose trust policy accepts the OIDC identity",
      "x-placeholder": "arn:aws:iam::123456789012:role/WebIdentityRole",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: ["gcp_web_identity", "web_identity_file"],
      },
    },
    aws_web_identity_role_session_name: {
      type: "string",
      title: "WebIdentity session name",
      description: "Optional session name for AssumeRoleWithWebIdentity",
      "x-placeholder": "rill-web-identity",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: ["gcp_web_identity", "web_identity_file"],
      },
      "x-advanced": true,
    },
    gcp_workload_identity_audience: {
      type: "string",
      title: "GCP workload identity audience",
      description:
        "Audience placed in the Google-signed OIDC token and matched by the AWS role trust policy",
      "x-placeholder": "rill-aws-access",
      "x-env-var-name": "GCP_WORKLOAD_IDENTITY_AUDIENCE",
      "x-step": "connector",
      "x-visible-if": { auth_method: "gcp_web_identity" },
    },
    path_prefixes: {
      type: "string",
      title: "Path prefixes",
      description:
        "Comma-separated list of bucket path prefixes this connector is allowed to access",
      "x-placeholder": "s3://my-bucket/path/",
      "x-step": "connector",
      "x-advanced": true,
    },
    allow_host_access: {
      type: "boolean",
      title: "Allow host access",
      description:
        "Use AWS credentials from the host environment (e.g. ~/.aws) in addition to configured credentials",
      "x-step": "connector",
      "x-advanced": true,
    },
    path: {
      type: "string",
      title: "S3 URI",
      description: "Path to your S3 bucket or prefix",
      pattern: "^s3://[^/]+(/.*)?$",
      errorMessage: {
        pattern: "Must be an S3 URI (e.g. s3://bucket/path)",
      },
      "x-placeholder": "s3://bucket/path",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_model",
      "x-step": "source",
    },
  },
  required: ["path", "name"],
  allOf: [
    {
      if: { properties: { auth_method: { const: "access_keys" } } },
      then: {
        required: ["aws_access_key_id", "aws_secret_access_key"],
      },
    },
    {
      if: { properties: { auth_method: { const: "gcp_web_identity" } } },
      then: {
        required: [
          "gcp_workload_identity_audience",
          "aws_web_identity_role_arn",
        ],
      },
    },
    {
      if: { properties: { auth_method: { const: "web_identity_file" } } },
      then: {
        required: ["aws_web_identity_token_file", "aws_web_identity_role_arn"],
      },
    },
  ],
};
