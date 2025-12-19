import * as yup from "yup";
import { dsnSchema, getYupSchema } from "./yupSchemas";
import { getConnectorSchema } from "./connector-schemas";
import {
  getFieldLabel,
  getRequiredFieldsByEnumValue,
} from "../../templates/schema-utils";
import type { AddDataFormType, MultiStepFormSchema } from "./types";

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
 * definitions alongside the schema utilities.
 */
function makeAuthOptionValidationSchema(
  connectorName: string,
  getAuthMethod?: () => string | undefined,
) {
  const schema = getConnectorSchema(connectorName);
  if (!schema) return null;

  const fieldValidations: Record<string, yup.StringSchema> = {};
  const requiredByMethod = getRequiredFieldsByEnumValue(schema, {
    step: "connector",
  });

  for (const [method, fields] of Object.entries(requiredByMethod || {})) {
    for (const fieldId of fields) {
      const label = getFieldLabel(schema as MultiStepFormSchema, fieldId);
      fieldValidations[fieldId] = (
        fieldValidations[fieldId] || yup.string()
      ).test(
        `required-${fieldId}-${method}`,
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
