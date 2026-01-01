import * as yup from "yup";

export const getYupSchema = {
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
};

export const dsnSchema = yup.object().shape({
  dsn: yup.string().required("DSN is required"),
});
