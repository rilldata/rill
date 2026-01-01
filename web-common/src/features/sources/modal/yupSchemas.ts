import * as yup from "yup";

export const getYupSchema: Record<string, yup.AnySchema> = {};

export const dsnSchema = yup.object().shape({
  dsn: yup.string().required("DSN is required"),
});
