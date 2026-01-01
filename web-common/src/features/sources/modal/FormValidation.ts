import type { ValidationAdapter } from "sveltekit-superforms/adapters";

import { createSchemasafeValidator } from "./jsonSchemaValidator";
import { getConnectorSchema } from "./connector-schemas";
import type { AddDataFormType } from "./types";

export function getValidationSchemaForConnector(
  name: string,
  formType: AddDataFormType,
): ValidationAdapter<Record<string, unknown>> {
  const jsonSchema = getConnectorSchema(name);
  const step = formType === "source" ? "source" : "connector";

  if (jsonSchema) return createSchemasafeValidator(jsonSchema, step);
  throw new Error(`No validation schema found for connector: ${name}`);
}
