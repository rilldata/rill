import type { MultiStepFormSchema } from "./types";

export const athenaSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    aws_access_key_id: {
      type: "string",
      title: "AWS Access Key ID",
      description: "AWS access key ID",
      "x-placeholder": "Enter AWS access key ID",
      "x-secret": true,
    },
    aws_secret_access_key: {
      type: "string",
      title: "AWS Secret Access Key",
      description: "AWS secret access key",
      "x-placeholder": "Enter AWS secret access key",
      "x-secret": true,
    },
    output_location: {
      type: "string",
      title: "S3 Output Location",
      description: "S3 URI for query results",
      "x-placeholder": "s3://my-bucket/athena-results/",
    },
    aws_role_arn: {
      type: "string",
      title: "IAM Role ARN (Optional)",
      description: "AWS IAM role ARN to assume (optional)",
      "x-placeholder": "arn:aws:iam::123456789012:role/MyRole",
    },
    region: {
      type: "string",
      title: "AWS Region",
      description: "AWS region where Athena is configured",
      "x-placeholder": "us-east-1",
    },
    workgroup: {
      type: "string",
      title: "Workgroup",
      description: "Athena workgroup name (optional)",
      "x-placeholder": "primary",
    },
  },
  required: ["aws_access_key_id", "aws_secret_access_key", "output_location"],
};
