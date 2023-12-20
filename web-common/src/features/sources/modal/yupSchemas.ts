import type { V1ConnectorSpec } from "@rilldata/web-common/runtime-client";
import * as yup from "yup";

export function getYupSchema(connector: V1ConnectorSpec) {
  switch (connector.name) {
    case "s3":
      return yup.object().shape({
        path: yup
          .string()
          .matches(/^s3:\/\//, "Must be an S3 URI (e.g. s3://bucket/path)")
          .required("S3 URI is required"),
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
        aws_region: yup.string(),
      });
    case "gcs":
      return yup.object().shape({
        path: yup
          .string()
          .matches(/^gs:\/\//, "Must be a GS URI (e.g. gs://bucket/path)")
          .required("GS URI is required"),
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
      });
    case "https":
      return yup.object().shape({
        path: yup
          .string()
          .matches(/^https?:\/\//, 'Path must start with "http(s)://"')
          .required("Path is required"),
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
      });
    case "duckdb":
      return yup.object().shape({
        sql: yup.string().required("sql is required"),
        db: yup.string().required("db is required"),
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
      });
    case "sqlite":
      return yup.object().shape({
        db: yup.string().required("db is required"),
        table: yup.string().required("table is required"),
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
      });
    case "bigquery":
      return yup.object().shape({
        sql: yup.string().required("sql is required"),
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
        project_id: yup.string().required("project_id is required"),
      });
    case "azure":
      return yup.object().shape({
        path: yup
          .string()
          .matches(
            /^azure:\/\//,
            "Must be an Azure URI (e.g. azure://container/path)"
          )
          .required("Path is required"),
        account: yup.string(),
      });
    case "postgres":
      return yup.object().shape({
        sql: yup.string().required("sql is required"),
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
        database_url: yup.string(),
      });
    case "snowflake":
      return yup.object().shape({
        sql: yup.string().required("sql is required"),
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
        dsn: yup.string(),
      });
    case "athena":
      return yup.object().shape({
        sql: yup.string().required("sql is required"),
        sourceName: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores"
          )
          .required("Source name is required"),
        output_location: yup.string(),
        workgroup: yup.string(),
      });
    default:
      throw new Error(`Unknown connector: ${connector.name}`);
  }
}

export function toYupFriendlyKey(key: string) {
  return key.replace(/\./g, "_");
}

export function fromYupFriendlyKey(key: string) {
  return key.replace(/_/g, ".");
}
