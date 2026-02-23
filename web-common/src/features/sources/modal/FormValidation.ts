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
 *
 * For multi-step connectors (e.g., S3), we initialize ALL fields regardless
 * of step so that source-step fields (like `path`) are tracked by superForm.
 * The renderer shows/hides fields based on the current step.
 */
export function createConnectorForm(args: {
  schemaName: string;
  formType: AddDataFormType;
  onUpdate: (event: { form: { data: FormData; valid: boolean } }) => void;
}) {
  const { schemaName, formType, onUpdate } = args;
  const schema = getConnectorSchema(schemaName);

  // Don't pass step filter - include defaults for ALL fields so multi-step
  // forms can track source/explorer fields even when starting on connector step
  const adapter = getValidationSchemaForConnector(schemaName, formType);

  // Get schema defaults (radio/tabs enums, explicit defaults)
  const schemaDefaults = schema ? getSchemaInitialValues(schema) : {};

  // Initialize ALL fields so superForm tracks them.
  // Without this, fields like `path` (no default) won't be in form.data on submit.
  const allFields: FormData = {};
  if (schema?.properties) {
    for (const [key, prop] of Object.entries(schema.properties)) {
      if (prop["x-display"] === "key-value") {
        allFields[key] = [];
      } else if (prop.type === "string") {
        allFields[key] = "";
      }
    }
  }

  // Merge: all fields as empty strings, then schema defaults on top
  const initialValues = { ...allFields, ...schemaDefaults };

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
