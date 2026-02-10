import type { MultiStepFormSchema } from "./types";

export const stagingSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Staging Import",
  "x-category": "staging",
  "x-form-height": "tall",
  "x-olap": {
    clickhouse: { formType: "source" },
  },
  properties: {
    warehouse: {
      type: "string",
      title: "Data warehouse",
      description: "Select an existing warehouse connector",
      "x-display": "select",
      "x-connector-drivers": ["snowflake", "redshift", "bigquery"],
    },
    sql: {
      type: "string",
      title: "SQL",
      description: "SQL query to run against your warehouse",
      "x-placeholder": "SELECT * FROM events",
    },
    staging_connector: {
      type: "string",
      title: "Staging connector",
      description:
        "Cloud storage used as temporary staging between the warehouse and ClickHouse",
      "x-display": "select",
      "x-connector-drivers": ["s3", "gcs", "azure"],
    },
    staging_path: {
      type: "string",
      title: "Staging path",
      description: "Cloud storage path for temporary staging data",
      "x-placeholder": "s3://bucket/temp-data",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_model",
    },
  },
  required: ["warehouse", "sql", "staging_connector", "staging_path", "name"],
};
