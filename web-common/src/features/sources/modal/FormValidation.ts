import type { ValidationAdapter } from "sveltekit-superforms/adapters";
import { superForm, defaults } from "sveltekit-superforms";

import { createSchemasafeValidator } from "./jsonSchemaValidator";
import { getConnectorSchema } from "./connector-schemas";
import type { AddDataFormType } from "./types";
import { getSchemaInitialValues } from "../../templates/schema-utils";

type FormData = Record<string, unknown>;

export function getValidationSchemaForConnector(
  name: string,
  formType: AddDataFormType,
  stepOverride?: "connector" | "source" | "explorer",
): ValidationAdapter<Record<string, unknown>> {
  const jsonSchema = getConnectorSchema(name);
  const step = stepOverride ?? (formType === "source" ? "source" : "connector");

  if (jsonSchema) return createSchemasafeValidator(jsonSchema, step);
  throw new Error(`No validation schema found for connector: ${name}`);
}

/**
 * Creates a SuperForm instance for a connector form, initialized from schema.
 * This is the single source of truth for form initialization.
 */
export function createConnectorForm(args: {
  schemaName: string;
  formType: AddDataFormType;
  onUpdate: (event: { form: { data: FormData; valid: boolean } }) => void;
}) {
  const { schemaName, formType, onUpdate } = args;
  const schema = getConnectorSchema(schemaName);
  const schemaStep = formType === "source" ? "source" : "connector";

  const adapter = getValidationSchemaForConnector(schemaName, formType);
  const initialValues = schema
    ? getSchemaInitialValues(schema, { step: schemaStep })
    : {};

  const formDefaults = defaults<FormData, string, FormData>(
    initialValues as Partial<FormData>,
    adapter,
  );

  return superForm<FormData, string, FormData>(formDefaults, {
    SPA: true,
    validators: adapter,
    onUpdate,
    resetForm: false,
    validationMethod: "onsubmit",
  });
}
