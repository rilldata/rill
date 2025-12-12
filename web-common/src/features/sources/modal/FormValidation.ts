import * as yup from "yup";
import { dsnSchema, getYupSchema } from "./yupSchemas";
import { multiStepFormConfigs } from "./multi-step-auth-configs";
import type { AddDataFormType, AuthField } from "./types";

export { dsnSchema };

export function getValidationSchemaForConnector(
  name: string,
  formType: AddDataFormType,
  opts?: {
    isMultiStepConnector?: boolean;
    authMethodGetter?: () => string | undefined;
  },
) {
  const { isMultiStepConnector, authMethodGetter } = opts || {};

  // For multi-step source flows, prefer the connector-specific schema when present
  // so step 1 (connector) validation doesn't require source-only fields.
  if (isMultiStepConnector && formType === "source") {
    const connectorKey = `${name}_connector`;
    if (connectorKey in getYupSchema) {
      return getYupSchema[connectorKey as keyof typeof getYupSchema];
    }
  }

  // For multi-step connector step, prefer connector-specific schema when present.
  if (isMultiStepConnector && formType === "connector") {
    // Generic dynamic schema based on auth options, driven by config.
    const dynamicSchema = makeAuthOptionValidationSchema(
      name,
      authMethodGetter,
    );
    if (dynamicSchema) return dynamicSchema;

    const connectorKey = `${name}_connector`;
    if (connectorKey in getYupSchema) {
      return getYupSchema[connectorKey as keyof typeof getYupSchema];
    }
  }

  return getYupSchema[name as keyof typeof getYupSchema];
}

/**
 * Build a yup schema that enforces required fields for the selected auth option
 * using the multi-step auth config. This keeps validation in sync with the UI
 * definitions in constants/multi-step-auth-configs.
 */
function makeAuthOptionValidationSchema(
  connectorName: string,
  getAuthMethod?: () => string | undefined,
) {
  const config =
    multiStepFormConfigs[connectorName as keyof typeof multiStepFormConfigs];
  if (!config) return null;

  // Collect all field definitions across auth methods.
  const fieldValidations: Record<string, yup.StringSchema> = {};

  for (const [method, fields] of Object.entries(config.authFieldGroups || {})) {
    for (const field of fields as AuthField[]) {
      // Only validate concrete input/credential fields.
      const required = !(field.optional ?? false);
      if (!required) continue;
      const label = field.type === "input" ? field.label || field.id : field.id;
      // Only apply requirement when the selected auth method matches.
      fieldValidations[field.id] = (
        fieldValidations[field.id] || yup.string()
      ).test(
        `required-${field.id}-${method}`,
        `${label} is required`,
        (value) => {
          if (!getAuthMethod) return true;
          const current = getAuthMethod();
          if (current !== method) return true;
          return !!value;
        },
      );
    }
  }

  // If nothing to validate, skip dynamic schema.
  if (!Object.keys(fieldValidations).length) return null;

  return yup.object().shape(fieldValidations);
}
