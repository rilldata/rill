import type { MultiStepFormSchema } from "./types";

export const redshiftSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    aws_access_key_id: {
      type: "string",
      title: "AWS Access Key ID",
      description: "AWS access key ID for Redshift Data API",
      "x-placeholder": "Enter AWS access key ID",
      "x-secret": true,
    },
    aws_secret_access_key: {
      type: "string",
      title: "AWS Secret Access Key",
      description: "AWS secret access key for Redshift Data API",
      "x-placeholder": "Enter AWS secret access key",
      "x-secret": true,
    },
    region: {
      type: "string",
      title: "AWS Region",
      description: "AWS region where the Redshift cluster is located",
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
      description: "Redshift Serverless workgroup name (for serverless)",
      "x-placeholder": "default-workgroup",
    },
    cluster_identifier: {
      type: "string",
      title: "Cluster Identifier",
      description: "Redshift provisioned cluster identifier (for provisioned clusters)",
      "x-placeholder": "my-redshift-cluster",
      "x-hint": "Provide either workgroup (for serverless) or cluster identifier (for provisioned)",
    },
  },
  required: ["aws_access_key_id", "aws_secret_access_key", "database"],
};
