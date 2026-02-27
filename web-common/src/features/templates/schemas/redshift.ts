import type { MultiStepFormSchema } from "./types";

export const redshiftSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Amazon Redshift",
  "x-category": "warehouse",
  properties: {
    aws_access_key_id: {
      type: "string",
      title: "AWS access key ID",
      description: "AWS access key ID",
      "x-placeholder": "your_access_key_id",
      "x-secret": true,
      "x-env-var-name": "AWS_ACCESS_KEY_ID",
    },
    aws_secret_access_key: {
      type: "string",
      title: "AWS secret access key",
      description: "AWS secret access key",
      "x-placeholder": "your_secret_access_key",
      "x-secret": true,
      "x-env-var-name": "AWS_SECRET_ACCESS_KEY",
    },
    region: {
      type: "string",
      title: "AWS region",
      description: "AWS region (e.g. us-east-1)",
      "x-placeholder": "us-east-1",
    },
    database: {
      type: "string",
      title: "Database",
      description: "Redshift database name",
      "x-placeholder": "dev",
    },
    workgroup: {
      type: "string",
      title: "Workgroup",
      description: "Redshift Serverless workgroup name",
      "x-placeholder": "default",
    },
    cluster_identifier: {
      type: "string",
      title: "Cluster identifier",
      description:
        "Redshift cluster identifier (use when not using serverless)",
      "x-placeholder": "redshift-cluster-1",
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
    "database",
    "sql",
    "name",
  ],
};
