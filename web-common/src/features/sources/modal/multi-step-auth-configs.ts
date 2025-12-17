import type {
  AuthOption,
  JSONSchemaConditional,
  MultiStepFormSchema,
} from "./types";
import { azureSchema } from "./schemas/azure";
import { gcsSchema } from "./schemas/gcs";
import { s3Schema } from "./schemas/s3";

type VisibleIf = Record<
  string,
  string | number | boolean | Array<string | number | boolean>
>;

export const multiStepFormSchemas: Record<string, MultiStepFormSchema> = {
  s3: s3Schema,
  gcs: gcsSchema,
  azure: azureSchema,
};

export function getConnectorSchema(
  connectorName: string,
): MultiStepFormSchema | null {
  const schema =
    multiStepFormSchemas[connectorName as keyof typeof multiStepFormSchemas];
  if (!schema?.properties) return null;
  return schema;
}

export function findAuthMethodKey(schema: MultiStepFormSchema): string | null {
  if (!schema.properties) return null;
  for (const [key, value] of Object.entries(schema.properties)) {
    if (value.enum && value["x-display"] === "radio") {
      return key;
    }
  }
  return schema.properties.auth_method ? "auth_method" : null;
}

export function getAuthOptionsFromSchema(
  schema: MultiStepFormSchema,
): { key: string; options: AuthOption[]; defaultMethod?: string } | null {
  const authMethodKey = findAuthMethodKey(schema);
  if (!authMethodKey) return null;
  const authProperty = schema.properties?.[authMethodKey];
  if (!authProperty?.enum) return null;

  const labels = authProperty["x-enum-labels"] ?? [];
  const descriptions = authProperty["x-enum-descriptions"] ?? [];
  const options =
    authProperty.enum?.map((value, idx) => ({
      value: String(value),
      label: labels[idx] ?? String(value),
      description:
        descriptions[idx] ?? authProperty.description ?? "Choose an option",
      hint: authProperty["x-hint"],
    })) ?? [];

  const defaultMethod =
    authProperty.default !== undefined && authProperty.default !== null
      ? String(authProperty.default)
      : options[0]?.value;

  return {
    key: authMethodKey,
    options,
    defaultMethod: defaultMethod || undefined,
  };
}

export function getRequiredFieldsByAuthMethod(
  schema: MultiStepFormSchema,
  opts?: { step?: "connector" | "source" },
): Record<string, string[]> {
  const authInfo = getAuthOptionsFromSchema(schema);
  if (!authInfo) return {};

  const conditionals = schema.allOf ?? [];
  const baseRequired = new Set(schema.required ?? []);
  const result: Record<string, string[]> = {};

  for (const option of authInfo.options) {
    const required = new Set<string>();

    // Start with base required fields.
    baseRequired.forEach((field) => {
      if (!opts?.step || isStepMatch(schema, field, opts.step)) {
        required.add(field);
      }
    });

    // Apply conditionals.
    for (const conditional of conditionals) {
      const matches = matchesAuthMethod(
        conditional,
        authInfo.key,
        option.value,
      );
      const target = matches ? conditional.then : conditional.else;
      target?.required?.forEach((field) => {
        if (!opts?.step || isStepMatch(schema, field, opts.step)) {
          required.add(field);
        }
      });
    }

    result[option.value] = Array.from(required);
  }

  return result;
}

export function getFieldLabel(
  schema: MultiStepFormSchema,
  key: string,
): string {
  return schema.properties?.[key]?.title || key;
}

export function isStepMatch(
  schema: MultiStepFormSchema,
  key: string,
  step: "connector" | "source",
): boolean {
  const prop = schema.properties?.[key];
  if (!prop) return false;
  return (prop["x-step"] ?? "connector") === step;
}

export function isVisibleForValues(
  schema: MultiStepFormSchema,
  key: string,
  values: Record<string, unknown>,
): boolean {
  const prop = schema.properties?.[key];
  if (!prop) return false;
  const conditions = prop["x-visible-if"];
  if (!conditions) return true;

  return Object.entries(conditions).every(([depKey, expected]) => {
    const actual = values?.[depKey];
    if (Array.isArray(expected)) {
      return expected.map(String).includes(String(actual));
    }
    return String(actual) === String(expected);
  });
}

function matchesAuthMethod(
  conditional: JSONSchemaConditional,
  authMethodKey: string,
  method: string,
) {
  const constValue =
    conditional.if?.properties?.[authMethodKey as keyof VisibleIf]?.const;
  if (constValue === undefined || constValue === null) return false;
  return String(constValue) === method;
}
