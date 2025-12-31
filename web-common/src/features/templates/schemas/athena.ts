import type { MultiStepFormSchema } from "./types";

export const athenaSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    aws_access_key_id: {
      type: "string",
      title: "AWS access key ID",
      description: "AWS access key ID used to authenticate to Athena",
      "x-placeholder": "your_access_key_id",
      "x-secret": true,
    },
    aws_secret_access_key: {
      type: "string",
      title: "AWS secret access key",
      description: "AWS secret access key paired with the access key ID",
      "x-placeholder": "your_secret_access_key",
      "x-secret": true,
    },
    output_location: {
      type: "string",
      title: "S3 output location",
      description:
        "S3 URI where Athena should write query results (e.g., s3://bucket/path/)",
      pattern: "^s3://.+",
      errorMessage: {
        pattern: "Must be an S3 URI (e.g., s3://bucket/path/)",
      },
      "x-placeholder": "s3://bucket-name/path/",
    },
  },
  required: ["aws_access_key_id", "aws_secret_access_key", "output_location"],
};

