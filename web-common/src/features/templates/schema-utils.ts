import type {
  JSONSchemaConditional,
  MultiStepFormSchema,
} from "./schemas/types";

export type RadioEnumOption = {
  value: string;
  label: string;
  description: string;
  hint?: string;
};

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

export function getFieldLabel(
  schema: MultiStepFormSchema,
  key: string,
): string {
  return schema.properties?.[key]?.title || key;
}

/**
 * Find all enum keys (radio or tabs) that have grouped fields.
 * These are UI control fields (like auth_method, connection_method) that shouldn't appear in YAML.
 */
export function findGroupedEnumKeys(schema: MultiStepFormSchema): string[] {
  if (!schema.properties) return [];
  const keys: string[] = [];
  for (const [key, value] of Object.entries(schema.properties)) {
    if (value.enum && value["x-grouped-fields"]) {
      keys.push(key);
    }
  }
  return keys;
}

export function findRadioEnumKey(schema: MultiStepFormSchema): string | null {
  if (!schema.properties) return null;
  for (const [key, value] of Object.entries(schema.properties)) {
    // Return radio or tabs fields that have grouped fields - those are auth/connection method selectors
    // Standalone radio fields (like "mode") should not be considered auth method keys
    const display = value["x-display"];
    if (
      value.enum &&
      (display === "radio" || display === "tabs") &&
      value["x-grouped-fields"]
    ) {
      return key;
    }
  }
  return schema.properties.auth_method ? "auth_method" : null;
}

export function getRadioEnumOptions(schema: MultiStepFormSchema): {
  key: string;
  options: RadioEnumOption[];
  defaultValue?: string;
} | null {
  const enumKey = findRadioEnumKey(schema);
  if (!enumKey) return null;
  const enumProperty = schema.properties?.[enumKey];
  if (!enumProperty?.enum) return null;

  const labels = enumProperty["x-enum-labels"] ?? [];
  const descriptions = enumProperty["x-enum-descriptions"] ?? [];
  const options =
    enumProperty.enum?.map((value, idx) => ({
      value: String(value),
      label: labels[idx] ?? String(value),
      description:
        descriptions[idx] ?? enumProperty.description ?? "Choose an option",
      hint: enumProperty["x-hint"],
    })) ?? [];

  const defaultValue =
    enumProperty.default !== undefined && enumProperty.default !== null
      ? String(enumProperty.default)
      : options[0]?.value;

  return {
    key: enumKey,
    options,
    defaultValue: defaultValue || undefined,
  };
}

export function getRequiredFieldsByEnumValue(
  schema: MultiStepFormSchema,
  opts?: { step?: "connector" | "source" | string },
): Record<string, string[]> {
  const enumInfo = getRadioEnumOptions(schema);
  if (!enumInfo) return {};

  const conditionals = schema.allOf ?? [];
  const baseRequired = new Set(schema.required ?? []);
  const result: Record<string, string[]> = {};

  const matchesStep = (field: string) => {
    if (!opts?.step) return true;
    const prop = schema.properties?.[field];
    if (!prop) return false;
    const propStep = prop["x-step"];
    if (!propStep) return true;
    return propStep === opts.step;
  };

  for (const option of enumInfo.options) {
    const required = new Set<string>();

    baseRequired.forEach((field) => {
      if (matchesStep(field)) {
        required.add(field);
      }
    });

    for (const conditional of conditionals) {
      const matches = matchesEnumCondition(
        conditional,
        enumInfo.key,
        option.value,
      );
      const target = matches ? conditional.then : conditional.else;
      target?.required?.forEach((field) => {
        if (matchesStep(field)) {
          required.add(field);
        }
      });
    }

    result[option.value] = Array.from(required);
  }

  return result;
}

function matchesEnumCondition(
  conditional: JSONSchemaConditional,
  enumKey: string,
  value: string,
) {
  const conditionProps = conditional.if?.properties;
  const constValue = conditionProps?.[enumKey]?.const;
  if (constValue === undefined || constValue === null) return false;
  return String(constValue) === value;
}
