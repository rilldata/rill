import * as yup from "yup";
import {
  INVALID_NAME_MESSAGE,
  VALID_NAME_PATTERN,
} from "../../entity-management/name-utils";

export const getYupSchema = {
  s3: yup.object().shape({
    path: yup
      .string()
      .matches(/^s3:\/\//, "Must be an S3 URI (e.g. s3://bucket/path)")
      .required("S3 URI is required"),
    aws_region: yup.string(),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  gcs: yup.object().shape({
    path: yup
      .string()
      .matches(/^gs:\/\//, "Must be a GS URI (e.g. gs://bucket/path)")
      .required("GS URI is required"),
  }),

  https: yup.object().shape({
    path: yup
      .string()
      .matches(/^https?:\/\//, 'Path must start with "http(s)://"')
      .required("Path is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  duckdb: yup.object().shape({
    path: yup.string().required("Path is required"),
  }),

  motherduck: yup.object().shape({
    dsn: yup.string().required("Connection string is required"),
    sql: yup.string().required("SQL is required"),
    token: yup.string().required("Access token is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  sqlite: yup.object().shape({
    db: yup.string().required("db is required"),
    table: yup.string().required("table is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  bigquery: yup.object().shape({
    google_application_credentials: yup.string().optional(),
    project_id: yup.string().optional(),
  }),

  azure: yup.object().shape({
    path: yup
      .string()
      .matches(
        /^azure:\/\//,
        "Must be an Azure URI (e.g. azure://container/path)",
      )
      .required("Path is required"),
    azure_storage_account: yup.string(),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  postgres: yup.object().shape({
    database_url: yup.string().required("Database URL is required"),
  }),

  snowflake: yup.object().shape({
    dsn: yup.string().optional(),
    account: yup.string().required("Account is required"),
    user: yup.string().required("Username is required"),
    password: yup.string().required("Password is required"),
    database: yup.string().required("Database is required"),
    schema: yup.string().optional(),
    warehouse: yup.string().optional(),
    role: yup.string().optional(),
  }),

  salesforce: yup.object().shape({
    soql: yup.string().required("soql is required"),
    sobject: yup.string().required("sobject is required"),
    name: yup
      .string()
      .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
      .required("Source name is required"),
  }),

  athena: yup.object().shape({
    aws_access_key_id: yup.string().required("AWS access key ID is required"),
    aws_secret_access_key: yup
      .string()
      .required("AWS secret access key is required"),
    output_location: yup.string().required("S3 URI is required"),
  }),

  redshift: yup.object().shape({
    aws_access_key_id: yup.string().required("AWS access key ID is required"),
    aws_secret_access_key: yup
      .string()
      .required("AWS secret access key is required"),
    workgroup: yup.string().optional(),
    region: yup.string().optional(), // TODO: add validation
    database: yup.string().required("database name is required"),
  }),

  mysql: yup.object().shape({
    dsn: yup.string().required("DSN is required"),
  }),

  clickhouse: yup.object().shape({
    managed: yup.boolean(),
    host: yup.string(),
    // .required("Host is required")
    // .matches(
    //   /^(?!https?:\/\/)[a-zA-Z0-9.-]+$/,
    //   "Do not prefix the host with `http(s)://`", // It will be added by the runtime
    // ),
    port: yup
      .string() // Purposefully using a string input, not a numeric input
      .matches(/^\d+$/, "Port must be a number"),
    username: yup.string(),
    password: yup.string(),
    cluster: yup.string(),
    ssl: yup.boolean(),
    name: yup.string(), // Required for typing
    // User-provided connector names requires a little refactor. Commenting out for now.
    // name: yup
    //   .string()
    //   .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
    //   .required("Connector name is required"),
  }),

  druid: yup.object().shape({
    host: yup
      .string()
      .required("Host is required")
      .matches(
        /^(?!https?:\/\/)[a-zA-Z0-9.-]+$/,
        "Do not prefix the host with `http(s)://`", // It will be added by the runtime
      ),
    port: yup
      .string() // Purposefully using a string input, not a numeric input
      .matches(/^\d+$/, "Port must be a number"),
    username: yup.string(),
    password: yup.string(),
    ssl: yup.boolean(),
    name: yup.string(), // Required for typing
    // User-provided connector names requires a little refactor. Commenting out for now.
    // name: yup
    //   .string()
    //   .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
    //   .required("Connector name is required"),
  }),

  pinot: yup.object().shape({
    broker_host: yup
      .string()
      .required("Broker host is required")
      .matches(
        /^(?!https?:\/\/)[a-zA-Z0-9.-]+$/,
        "Do not prefix the host with `http(s)://`", // It will be added by the runtime
      ),
    broker_port: yup
      .string() // Purposefully using a string input, not a numeric input
      .matches(/^\d+$/, "Port must be a number"),
    controller_host: yup
      .string()
      .required("Controller host is required")
      .matches(
        /^(?!https?:\/\/)[a-zA-Z0-9.-]+$/,
        "Do not prefix the host with `http(s)://`", // It will be added by the runtime
      ),
    controller_port: yup
      .string() // Purposefully using a string input, not a numeric input
      .matches(/^\d+$/, "Port must be a number"),
    username: yup.string(),
    password: yup.string(),
    ssl: yup.boolean(),
    name: yup.string(), // Required for typing
    // User-provided connector names requires a little refactor. Commenting out for now.
    // name: yup
    //   .string()
    //   .matches(VALID_NAME_PATTERN, INVALID_NAME_MESSAGE)
    //   .required("Connector name is required"),
  }),
};

export const dsnSchema = yup.object().shape({
  dsn: yup.string().required("DSN is required"),
});
