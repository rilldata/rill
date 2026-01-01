import * as yup from "yup";

export const getYupSchema = {
  duckdb: yup.object().shape({
    path: yup.string().required("path is required"),
    attach: yup.string().optional(),
  }),

  motherduck: yup.object().shape({
    token: yup.string().required("Token is required"),
    path: yup.string().required("Path is required"),
    schema_name: yup.string().required("Schema name is required"),
  }),

  clickhouse: yup.object().shape({
    dsn: yup.string().optional(),
    managed: yup.boolean(),
    host: yup.string(),
    port: yup
      .string() // Purposefully using a string input, not a numeric input
      .matches(/^\d+$/, "Port must be a number"),
    username: yup.string(),
    password: yup.string(),
    cluster: yup.string(),
    ssl: yup.boolean(),
    name: yup.string(), // Required for typing
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
  }),
};

export const dsnSchema = yup.object().shape({
  dsn: yup.string().required("DSN is required"),
});
