import type { ValidationAdapter } from "sveltekit-superforms/adapters";
import { yup as yupAdapter } from "sveltekit-superforms/adapters";

import { createSchemasafeValidator } from "./jsonSchemaValidator";
import { getConnectorSchema } from "./connector-schemas";
import { dsnSchema, getYupSchema } from "./yupSchemas";
import type { AddDataFormType } from "./types";

export { dsnSchema };

export function getValidationSchemaForConnector(
  name: string,
  formType: AddDataFormType,
  opts?: {
    isMultiStepConnector?: boolean;
  },
): ValidationAdapter<Record<string, unknown>> {
  const { isMultiStepConnector } = opts || {};
  const jsonSchema = getConnectorSchema(name);

  if (isMultiStepConnector && jsonSchema) {
    const step = formType === "source" ? "source" : "connector";
    return createSchemasafeValidator(jsonSchema, step);
  }

  const fallbackYupSchema = getYupSchema[name as keyof typeof getYupSchema];
  return yupAdapter(fallbackYupSchema);
}

