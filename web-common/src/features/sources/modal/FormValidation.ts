import { dsnSchema, getYupSchema } from "./yupSchemas";
import type { AddDataFormType } from "./types";

export { dsnSchema };

export function getValidationSchemaForConnector(
  name: string,
  formType: AddDataFormType,
  opts?: { isMultiStepConnector?: boolean },
) {
  const { isMultiStepConnector } = opts || {};

  // For multi-step source flows, prefer the connector-specific schema when present
  // so step 1 (connector) validation doesn't require source-only fields.
  if (isMultiStepConnector && formType === "source") {
    const connectorKey = `${name}_connector`;
    if (connectorKey in getYupSchema) {
      return getYupSchema[connectorKey as keyof typeof getYupSchema];
    }
  }

  return getYupSchema[name as keyof typeof getYupSchema];
}
