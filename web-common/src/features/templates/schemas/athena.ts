import type { MultiStepFormSchema } from "./types";

export const athenaSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Amazon Athena",
  "x-category": "warehouse",
  "x-olap": {
    duckdb: { formType: "connector" },
  },
  properties: {
    aws_access_key_id: {
      type: "string",
      title: "AWS access key ID",
      description: "AWS access key ID used to authenticate to Athena",
      "x-placeholder": "your_access_key_id",
      "x-secret": true,
      "x-env-var-name": "AWS_ACCESS_KEY_ID",
      "x-step": "connector",
    },
    aws_secret_access_key: {
      type: "string",
      title: "AWS secret access key",
      description: "AWS secret access key paired with the access key ID",
      "x-placeholder": "your_secret_access_key",
      "x-secret": true,
      "x-env-var-name": "AWS_SECRET_ACCESS_KEY",
      "x-step": "connector",
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
      "x-step": "connector",
    },
    sql: {
      type: "string",
      title: "SQL",
      description: "SQL query to run against your warehouse",
      "x-placeholder": "Input SQL",
      "x-step": "explorer",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_model",
      "x-step": "explorer",
    },
  },
  required: [
    "aws_access_key_id",
    "aws_secret_access_key",
    "output_location",
    "sql",
    "name",
  ],
};
