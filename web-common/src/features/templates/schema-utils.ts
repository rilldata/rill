import type {
  JSONSchemaConditional,
  JSONSchemaField,
  MultiStepFormSchema,
} from "./schemas/types";

export type RadioEnumOption = {
  value: string;
  label: string;
  description: string;
  hint?: string;
};

export function isStepMatch(
  schema: MultiStepFormSchema | null,
  key: string,
  step?: "connector" | "source" | string,
): boolean {
  if (!schema?.properties) return false;
  const prop = schema.properties[key];
  if (!prop) return false;
  if (!step) return true;
  const propStep = prop["x-step"];
  if (!propStep) return step !== "explorer";
  return propStep === step;
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

export function getFieldLabel(
  schema: MultiStepFormSchema,
  key: string,
): string {
  return schema.properties?.[key]?.title || key;
}

export type SchemaFieldMeta = {
  key: string;
  type?: "string" | "number" | "boolean" | "object";
  displayName: string;
  description?: string;
  placeholder?: string;
  hint?: string;
  secret?: boolean;
  docsUrl?: string;
  required?: boolean;
  default?: string | number | boolean;
  enum?: Array<string | number | boolean>;
  informational?: boolean;
  internal?: boolean;
};

export function getSchemaFieldMetaList(
  schema: MultiStepFormSchema,
  opts?: { step?: "connector" | "source" | string },
): SchemaFieldMeta[] {
  const properties = schema.properties ?? {};
  const required = new Set<string>(
    (schema.required ?? []).filter((key) =>
      isStepMatch(schema, key, opts?.step),
    ),
  );

  return Object.entries(properties)
    .filter(([key]) => isStepMatch(schema, key, opts?.step))
    .map(([key, prop]) => ({
      key,
      type: prop.type,
      displayName: prop.title ?? key,
      description: prop.description,
      placeholder: prop["x-placeholder"],
      hint: prop["x-hint"],
      secret: Boolean(prop["x-secret"]),
      docsUrl: prop["x-docs-url"],
      required: required.has(key),
      default: prop.default,
      enum: prop.enum,
      informational: Boolean(prop["x-informational"]),
      internal: Boolean(prop["x-internal"]),
    }));
}

export function getSchemaInitialValues(
  schema: MultiStepFormSchema,
  opts?: { step?: "connector" | "source" | string },
): Record<string, unknown> {
  const initial: Record<string, unknown> = {};
  const properties = schema.properties ?? {};

  for (const [key, prop] of Object.entries(properties)) {
    if (!isStepMatch(schema, key, opts?.step)) continue;
    if (prop.default !== undefined && prop.default !== null) {
      initial[key] = prop.default;
      continue;
    }
    if (
      prop.enum?.length &&
      (prop["x-display"] === "radio" || prop["x-display"] === "tabs")
    ) {
      initial[key] = String(prop.enum[0]);
    }
  }

  return initial;
}

export function getRequiredFieldsForValues(
  schema: MultiStepFormSchema,
  values: Record<string, unknown>,
  step?: "connector" | "source" | string,
): Set<string> {
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

  return required;
}

export function getSchemaSecretKeys(
  schema: MultiStepFormSchema,
  opts?: { step?: "connector" | "source" | string },
): string[] {
  const properties = schema.properties ?? {};
  return Object.entries(properties)
    .filter(
      ([key, prop]) =>
        isStepMatch(schema, key, opts?.step) && Boolean(prop["x-secret"]),
    )
    .map(([key]) => key);
}

export function getSchemaStringKeys(
  schema: MultiStepFormSchema,
  opts?: { step?: "connector" | "source" | string },
): string[] {
  const properties = schema.properties ?? {};
  return Object.entries(properties)
    .filter(
      ([key, prop]) =>
        isStepMatch(schema, key, opts?.step) && prop.type === "string",
    )
    .map(([key]) => key);
}

export function filterSchemaInternalValues(
  schema: MultiStepFormSchema,
  values: Record<string, unknown>,
  opts?: { step?: "connector" | "source" | string },
): Record<string, unknown> {
  const properties = schema.properties ?? {};
  return Object.fromEntries(
    Object.entries(values).filter(([key]) => {
      const prop = properties[key] as JSONSchemaField | undefined;
      if (!prop) return false;
      if (!isStepMatch(schema, key, opts?.step)) return false;
      return !prop["x-internal"];
    }),
  );
}

export function filterSchemaValuesForSubmit(
  schema: MultiStepFormSchema,
  values: Record<string, unknown>,
  opts?: { step?: "connector" | "source" | string },
): Record<string, unknown> {
  const tabFiltered = filterValuesByTabGroups(schema, values, opts);
  return filterSchemaInternalValues(schema, tabFiltered, opts);
}

export function findRadioEnumKey(schema: MultiStepFormSchema): string | null {
  if (!schema.properties) return null;
  for (const [key, value] of Object.entries(schema.properties)) {
    if (value.enum && value["x-display"] === "radio") {
      return key;
    }
  }
  return null;
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

  const hasDefault =
    enumProperty.default !== undefined && enumProperty.default !== null;
  const defaultValue = hasDefault
    ? String(enumProperty.default)
    : options[0]?.value;

  return {
    key: enumKey,
    options,
    defaultValue,
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

function filterValuesByTabGroups(
  schema: MultiStepFormSchema,
  values: Record<string, unknown>,
  opts?: { step?: "connector" | "source" | string },
) {
  const properties = schema.properties ?? {};
  const result = { ...values };

  for (const [key, prop] of Object.entries(properties)) {
    if (!isStepMatch(schema, key, opts?.step)) continue;
    if (prop["x-display"] !== "tabs") continue;
    const tabGroups = prop["x-tab-group"];
    if (!tabGroups) continue;
    const selected = String(values?.[key] ?? "");
    const active = tabGroups[selected] ?? [];
    const allChildKeys = new Set(Object.values(tabGroups).flat());
    for (const childKey of allChildKeys) {
      if (active.includes(childKey)) continue;
      delete result[childKey];
    }
  }

  return result;
}

/**
 * Returns values that should be enforced based on allOf/if/then conditionals.
 * This includes:
 * - `const` values from matching conditional branches (must be enforced)
 * - `default` values from matching conditional branches (applied when field is empty)
 */
export function getConditionalValues(
  schema: MultiStepFormSchema,
  values: Record<string, unknown>,
): Record<string, unknown> {
  const result: Record<string, unknown> = {};

  for (const conditional of schema.allOf ?? []) {
    const condition = conditional.if?.properties;
    const matches = matchesCondition(condition, values);
    const branch = matches ? conditional.then : conditional.else;

    if (!branch?.properties) continue;

    for (const [key, prop] of Object.entries(branch.properties)) {
      // Enforce const values - these must always be applied
      if (prop.const !== undefined) {
        result[key] = prop.const;
      }
      // Apply default values only when current value is empty/unset
      else if (
        prop.default !== undefined &&
        (values[key] === undefined ||
          values[key] === null ||
          values[key] === "")
      ) {
        result[key] = prop.default;
      }
    }
  }

  return result;
}
