import type {
  JSONSchemaConditional,
  JSONSchemaField,
  MultiStepFormSchema,
} from "./types";

type Step = string | null | undefined;

export function matchesStep(prop: JSONSchemaField | undefined, step: Step) {
  if (!step) return true;
  const fieldStep = prop?.["x-step"];
  return fieldStep ? fieldStep === step : true;
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

export function visibleFieldsForValues(
  schema: MultiStepFormSchema,
  values: Record<string, unknown>,
  step?: Step,
): Array<[string, JSONSchemaField]> {
  const properties = schema.properties ?? {};
  return Object.entries(properties).filter(([key, prop]) => {
    if (!matchesStep(prop, step)) return false;
    return isVisibleForValues(schema, key, values);
  });
}

export function computeRequiredFields(
  schema: MultiStepFormSchema,
  values: Record<string, unknown>,
  step?: Step,
): Set<string> {
  const required = new Set<string>();
  const properties = schema.properties ?? {};

  // Base required fields.
  for (const field of schema.required ?? []) {
    if (!step || matchesStep(properties[field], step)) {
      required.add(field);
    }
  }

  // Conditional required fields driven by `allOf`.
  for (const conditional of schema.allOf ?? []) {
    const applies = matchesConditional(conditional, values);
    const target = applies ? conditional.then : conditional.else;
    for (const field of target?.required ?? []) {
      if (!step || matchesStep(properties[field], step)) {
        required.add(field);
      }
    }
  }

  return required;
}

export function dependsOnField(prop: JSONSchemaField, dependency: string) {
  const conditions = prop["x-visible-if"];
  if (!conditions) return false;
  return Object.prototype.hasOwnProperty.call(conditions, dependency);
}

export function keysDependingOn(
  schema: MultiStepFormSchema,
  dependencies: string[],
  step?: Step,
): Set<string> {
  const properties = schema.properties ?? {};
  const result = new Set<string>();

  for (const [key, prop] of Object.entries(properties)) {
    if (!matchesStep(prop, step)) continue;
    if (dependencies.some((dep) => dependsOnField(prop, dep))) {
      result.add(key);
    }
  }

  return result;
}

function matchesConditional(
  conditional: JSONSchemaConditional,
  values: Record<string, unknown>,
) {
  const conditions = conditional.if?.properties;
  if (!conditions) return false;

  return Object.entries(conditions).every(([depKey, constraint]) => {
    if (!("const" in constraint)) return false;
    const actual = values?.[depKey];
    return String(actual) === String(constraint.const);
  });
}
