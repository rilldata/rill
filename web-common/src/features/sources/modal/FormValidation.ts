import * as yup from "yup";
import { dsnSchema, getYupSchema } from "./yupSchemas";
import { getConnectorSchema } from "./connector-schemas";
import {
  findRadioEnumKey,
  getFieldLabel,
  getRadioEnumOptions,
  getRequiredFieldsByEnumValue,
} from "../../templates/schema-utils";
import { isEmpty } from "./utils";
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
    return withJsonSchemaAuthValidation(name, authMethodGetter);
  }

  return getYupSchema[name as keyof typeof getYupSchema];
}

function withJsonSchemaAuthValidation(
  connectorName: string,
  getAuthMethod?: () => string | undefined,
) {
  const connectorKey =
    `${connectorName}_connector` as keyof typeof getYupSchema;
  const baseSchema =
    (getYupSchema[connectorKey] as yup.ObjectSchema<any> | undefined) ||
    yup.object();
  const schema = getConnectorSchema(connectorName);
  if (!schema) return baseSchema;

  return baseSchema.test(
    "auth-required-fields",
    "Missing required fields for selected auth method",
    function (value) {
      const values = (value || {}) as Record<string, unknown>;
      const authMethod = resolveAuthMethod(schema, values, getAuthMethod);
      if (!authMethod) return true;

      const requiredByMethod = getRequiredFieldsByEnumValue(schema, {
        step: "connector",
      });
      const requiredFields = requiredByMethod[authMethod] ?? [];

      const errors = requiredFields
        .filter((fieldId) => isEmpty(values[fieldId]))
        .map((fieldId) =>
          this.createError({
            path: fieldId,
            message: `${getFieldLabel(schema, fieldId)} is required`,
          }),
        );

      if (!errors.length) return true;
      return new yup.ValidationError(errors);
    },
  );
}

function resolveAuthMethod(
  schema: MultiStepFormSchema,
  values: Record<string, unknown>,
  getAuthMethod?: () => string | undefined,
) {
  const authInfo = getRadioEnumOptions(schema);
  const options = authInfo?.options ?? [];
  const normalize = (method?: string | null) =>
    method && options.some((opt) => opt.value === method) ? method : undefined;

  const fromGetter = normalize(getAuthMethod?.());
  const authKey = authInfo?.key || findRadioEnumKey(schema);
  const rawFromForm =
    authKey && values?.[authKey] != null ? String(values[authKey]) : undefined;
  const fromForm = normalize(rawFromForm);
  const fallback = normalize(authInfo?.defaultValue) || options[0]?.value;

  return fromGetter || fromForm || fallback;
}
