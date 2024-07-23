import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import * as yup from "yup";
import {
  INVALID_NAME_MESSAGE,
  VALID_NAME_PATTERN,
} from "../../entity-management/name-utils";

export function getYupSchema(connector: V1ConnectorDriver) {
  switch (connector.name) {
    case "s3":
      return yup.object().shape({
        path: yup
          .string()
          .matches(/^s3:\/\//, "Must be an S3 URI (e.g. s3://bucket/path)")
          .required("S3 URI is required"),
        aws_region: yup.string(),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "gcs":
      return yup.object().shape({
        path: yup
          .string()
          .matches(/^gs:\/\//, "Must be a GS URI (e.g. gs://bucket/path)")
          .required("GS URI is required"),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "https":
      return yup.object().shape({
        path: yup
          .string()
          .matches(/^https?:\/\//, 'Path must start with "http(s)://"')
          .required("Path is required"),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "duckdb":
      return yup.object().shape({
        db: yup.string().required("db is required"),
        sql: yup.string().required("sql is required"),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "motherduck":
      return yup.object().shape({
        connection_string: yup
          .string()
          .required("Connection string is required"),
        sql: yup.string().required("SQL is required"),
        token: yup.string().required("Access token is required"),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "sqlite":
      return yup.object().shape({
        db: yup.string().required("db is required"),
        table: yup.string().required("table is required"),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "bigquery":
      return yup.object().shape({
        sql: yup.string().required("sql is required"),
        project_id: yup.string().required("project_id is required"),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "azure":
      return yup.object().shape({
        path: yup
          .string()
          .matches(
            /^azure:\/\//,
            "Must be an Azure URI (e.g. azure://container/path)",
          )
          .required("Path is required"),
        account: yup.string(),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "postgres":
      return yup.object().shape({
        sql: yup.string().required("sql is required"),
        database_url: yup.string(),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "snowflake":
      return yup.object().shape({
        sql: yup.string().required("sql is required"),
        dsn: yup.string(),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "salesforce":
      return yup.object().shape({
        soql: yup.string().required("soql is required"),
        sobject: yup.string().required("sobject is required"),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "athena":
      return yup.object().shape({
        sql: yup.string().required("sql is required"),
        output_location: yup.string(),
        workgroup: yup.string(),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "redshift":
      return yup.object().shape({
        sql: yup.string().required("SQL is required"),
        database: yup.string().required("database name is required"),
        output_location: yup
          .string()
          .required("S3 location for temporary files"),
        workgroup: yup.string().optional(),
        cluster_identifier: yup.string().optional(),
        role_arn: yup
          .string()
          .required("Role ARN associated with the Redshift cluster"),
        region: yup.string().optional(),
        name: yup
          .string()
          .matches(
            /^[a-zA-Z_][a-zA-Z0-9_]*$/,
            "Source name must start with a letter or underscore and contain only letters, numbers, and underscores",
          )
          .required("Source name is required"),
      });

    case "mysql":
      return yup.object().shape({
        sql: yup.string().required("sql is required"),
        dsn: yup.string(),
        name: yup
          .string()
          .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
          .required("Source name is required"),
      });

    case "clickhouse":
      return yup.object().shape({
        host: yup
          .string()
          .required("Host is required")
          .matches(
            /^(?!https?:\/\/)[a-zA-Z0-9.-]+$/,
            "Do not prefix the host with `http(s)://`", // It will be added by the runtime
          ),
        port: yup.number().required("Port is required"),
        username: yup.string(),
        password: yup.string(),
        ssl: yup.boolean(),
        // User-provided connector names requires a little refactor. Commenting out for now.
        // name: yup
        //   .string()
        //   .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
        //   .required("Connector name is required"),
      });

    case "druid":
      return yup.object().shape({
        host: yup
          .string()
          .required("Host is required")
          .matches(
            /^(?!https?:\/\/)[a-zA-Z0-9.-]+$/,
            "Do not prefix the host with `http(s)://`", // It will be added by the runtime
          ),
        port: yup.number().required("Port is required"),
        username: yup.string(),
        password: yup.string(),
        ssl: yup.boolean(),
        // User-provided connector names requires a little refactor. Commenting out for now.
        // name: yup
        //   .string()
        //   .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
        //   .required("Connector name is required"),
      });

    case "pinot":
      return yup.object().shape({
        host: yup
          .string()
          .required("Host is required")
          .matches(
            /^(?!https?:\/\/)[a-zA-Z0-9.-]+$/,
            "Do not prefix the host with `http(s)://`", // It will be added by the runtime
          ),
        port: yup.number().required("Port is required"),
        username: yup.string(),
        password: yup.string(),
        ssl: yup.boolean(),
        // User-provided connector names requires a little refactor. Commenting out for now.
        // name: yup
        //   .string()
        //   .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
        //   .required("Connector name is required"),
      });

    default:
      return yup.object().shape({});
  }
}

export function toYupFriendlyKey(key: string) {
  return key.replace(/\./g, "_");
}

export function fromYupFriendlyKey(key: string) {
  return key.replace(/_/g, ".");
}
