import * as yup from "yup";
import { dsnSchema, getYupSchema } from "./yupSchemas";
import { getConnectorSchema } from "./connector-schemas";
import {
  findRadioEnumKey,
  getFieldLabel,
  getRadioEnumOptions,
  getRequiredFieldsByEnumValue,
  isStepMatch,
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
    return withJsonSchemaSourceValidation(name);
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
  // Avoid relying on connector-specific Yup schemas for multi-step flows;
  // defer entirely to JSON schema-driven validation.
  const baseSchema = yup.object();
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

function withJsonSchemaSourceValidation(connectorName: string) {
  // Source step validation should also be driven by JSON schema definitions.
  const baseSchema = yup.object();
  const schema = getConnectorSchema(connectorName);
  if (!schema) return baseSchema;

  return baseSchema.test(
    "source-required-fields",
    "Missing required fields for source step",
    function (value) {
      const values = (value || {}) as Record<string, unknown>;
      const errors: yup.ValidationError[] = [];
      const requiredFields = getRequiredFieldsForStep(schema, values, "source");

      errors.push(
        ...requiredFields
          .filter((fieldId) => isEmpty(values[fieldId]))
          .map((fieldId) =>
            this.createError({
              path: fieldId,
              message: `${getFieldLabel(schema, fieldId)} is required`,
            }),
          ),
      );

      const patternValidators = getPatternValidatorsForStep(schema, "source");
      for (const validator of patternValidators) {
        const raw = values[validator.key];
        if (isEmpty(raw)) continue;
        const stringValue = String(raw);

        if (validator.pattern && !validator.pattern.test(stringValue)) {
          errors.push(
            this.createError({
              path: validator.key,
              message: patternErrorMessage(schema, validator.key),
            }),
          );
          continue;
        }

        if (
          validator.format &&
          !isFormatValid(validator.format, stringValue)
        ) {
          errors.push(
            this.createError({
              path: validator.key,
              message: formatErrorMessage(
                schema,
                validator.key,
                validator.format,
              ),
            }),
          );
        }
      }

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

function getRequiredFieldsForStep(
  schema: MultiStepFormSchema,
  values: Record<string, unknown>,
  step: "connector" | "source" | string,
): string[] {
  const required = new Set<string>();
  (schema.required ?? []).forEach((key) => {
    if (isStepMatch(schema, key, step)) required.add(key);
  });

  for (const conditional of schema.allOf ?? []) {
    const condition = conditional.if?.properties;
    const matches = matchesCondition(condition, values);
    const branch = matches ? conditional.then : conditional.else;
    branch?.required?.forEach((key) => {
      if (isStepMatch(schema, key, step)) required.add(key);
    });
  }

  return Array.from(required);
}

function matchesCondition(
  condition: Record<string, { const?: string | number | boolean }> | undefined,
  values: Record<string, unknown>,
) {
  if (!condition || !Object.keys(condition).length) return false;
  return Object.entries(condition).every(([depKey, def]) => {
    if (def.const === undefined || def.const === null) return false;
    return String(values?.[depKey]) === String(def.const);
  });
}

function getPatternValidatorsForStep(
  schema: MultiStepFormSchema,
  step: "connector" | "source" | string,
) {
  const props = schema.properties ?? {};
  return Object.entries(props)
    .filter(([key]) => isStepMatch(schema, key, step))
    .map(([key, prop]) => ({
      key,
      pattern: compilePattern((prop as any).pattern),
      format: (prop as any).format as string | undefined,
      errorMessage: ((prop as any).errorMessage || {}) as Record<
        string,
        string
      >,
    }))
    .filter((entry) => entry.pattern || entry.format);
}

function compilePattern(pattern?: string | RegExp | null): RegExp | null {
  if (!pattern) return null;
  if (pattern instanceof RegExp) return pattern;
  try {
    return new RegExp(pattern);
  } catch {
    return null;
  }
}

function isFormatValid(format: string, value: string): boolean {
  if (format === "uri") {
    try {
      // URL supports custom schemes such as s3://
      new URL(value);
      return true;
    } catch {
      return false;
    }
  }
  return true;
}

function patternErrorMessage(schema: MultiStepFormSchema, key: string): string {
  const prop = schema.properties?.[key] as any;
  const custom = prop?.errorMessage?.pattern as string | undefined;
  if (custom) return custom;
  const label = getFieldLabel(schema, key);
  return `${label} is invalid`;
}

function formatErrorMessage(
  schema: MultiStepFormSchema,
  key: string,
  format: string,
): string {
  const prop = schema.properties?.[key] as any;
  const custom = prop?.errorMessage?.format as string | undefined;
  if (custom) return custom;
  const label = getFieldLabel(schema, key);
  return `${label} must be a valid ${format}`;
}
