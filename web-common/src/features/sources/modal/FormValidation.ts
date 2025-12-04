import { dsnSchema, getYupSchema } from "./yupSchemas";

export { dsnSchema };

export function getValidationSchemaForConnector(name: string) {
  return getYupSchema[name as keyof typeof getYupSchema];
}
