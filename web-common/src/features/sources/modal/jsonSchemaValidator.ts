import type { ValidatorOptions, ValidationError } from "@exodus/schemasafe";
import { validator as compileValidator } from "@exodus/schemasafe";
import { schemasafe } from "sveltekit-superforms/adapters";
import type { ValidationAdapter } from "sveltekit-superforms/adapters";

import { getFieldLabel, isStepMatch } from "../../templates/schema-utils";
import type {
  JSONSchemaConditional,
  JSONSchemaConstraint,
  MultiStepFormSchema,
} from "../../templates/schemas/types";

const DEFAULT_SCHEMASAFE_OPTIONS: ValidatorOptions = {
  includeErrors: true,
  allErrors: true,
  allowUnusedKeywords: true,
  formats: {
    uri: (value: string) => {
      if (typeof value !== "string") return false;
      try {
        // Allow custom schemes such as s3:// or gs://
        new URL(value);
        return true;
      } catch {
        return false;
      }
    },
    // We treat file inputs as strings; superforms handles the upload.
    file: () => true,
  },
};

type Step = "connector" | "source" | string | undefined;

export function buildStepSchema(
  schema: MultiStepFormSchema,
  step: Step,
): MultiStepFormSchema {
  const properties = Object.entries(schema.properties ?? {}).reduce<
    NonNullable<MultiStepFormSchema["properties"]>
  >((acc, [key, prop]) => {
    if (!isStepMatch(schema, key, step)) return acc;
    acc[key] = prop;
    return acc;
  }, {});

  const required = (schema.required ?? []).filter((key) =>
    isStepMatch(schema, key, step),
  );

  const filteredAllOf = (schema.allOf ?? [])
    .map((conditional) => filterConditional(conditional, schema, step))
    .filter((conditional): conditional is JSONSchemaConditional =>
      Boolean(
        conditional &&
          ((conditional.then?.required?.length ?? 0) > 0 ||
            (conditional.else?.required?.length ?? 0) > 0),
      ),
    );

  return {
    $schema: schema.$schema,
    type: "object",
    properties,
    ...(required.length ? { required } : {}),
    ...(filteredAllOf.length ? { allOf: filteredAllOf } : {}),
  };
}

export function createSchemasafeValidator(
  schema: MultiStepFormSchema,
  step: Step,
  opts?: { config?: ValidatorOptions },
): ValidationAdapter<Record<string, unknown>> {
  const stepSchema = buildStepSchema(schema, step);
  const validator = compileValidator(stepSchema, {
    ...DEFAULT_SCHEMASAFE_OPTIONS,
    ...opts?.config,
  });

  const baseAdapter = schemasafe(stepSchema, {
    config: {
      ...DEFAULT_SCHEMASAFE_OPTIONS,
      ...opts?.config,
    },
  });

  return {
    ...baseAdapter,
    async validate(data: Record<string, unknown> = {}) {
      const pruned = pruneEmptyFields(data);
      const isValid = validator(pruned as any);
      if (isValid) {
        return { data: pruned, success: true };
      }

      const issues = (validator.errors ?? []).map((error) =>
        toIssue(error, schema),
      );
      return { success: false, issues };
    },
  };
}

function pruneEmptyFields(
  values: Record<string, unknown>,
): Record<string, unknown> {
  const result: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(values ?? {})) {
    if (value === "" || value === null || value === undefined) continue;
    result[key] = value;
  }
  return result;
}

function filterConditional(
  conditional: JSONSchemaConditional,
  schema: MultiStepFormSchema,
  step: Step,
): JSONSchemaConditional | null {
  const thenRequired = filterRequired(conditional.then, schema, step);
  const elseRequired = filterRequired(conditional.else, schema, step);

  if (!thenRequired.length && !elseRequired.length) return null;

  return {
    if: conditional.if,
    then: thenRequired.length ? { required: thenRequired } : undefined,
    else: elseRequired.length ? { required: elseRequired } : undefined,
  };
}

function filterRequired(
  constraint: JSONSchemaConstraint | undefined,
  schema: MultiStepFormSchema,
  step: Step,
): string[] {
  return (constraint?.required ?? []).filter((key) =>
    isStepMatch(schema, key, step),
  );
}

function toIssue(error: ValidationError, schema: MultiStepFormSchema) {
  const pathSegments = parseInstanceLocation(error.instanceLocation);
  const key = pathSegments[0];
  return {
    path: pathSegments,
    message: buildMessage(schema, key, error),
  };
}

function buildMessage(
  schema: MultiStepFormSchema,
  key: string | undefined,
  error: ValidationError,
): string {
  if (!key) return "Invalid value";
  const prop = schema.properties?.[key] as any;
  const label = getFieldLabel(schema, key);
  const keyword = parseKeyword(error.keywordLocation);

  if (keyword === "required") return `${label} is required`;

  if (keyword === "pattern") {
    const custom = prop?.errorMessage?.pattern as string | undefined;
    return custom || `${label} is invalid`;
  }

  if (keyword === "format") {
    const custom = prop?.errorMessage?.format as string | undefined;
    const format = prop?.format as string | undefined;
    if (custom) return custom;
    if (format) return `${label} must be a valid ${format}`;
    return `${label} is invalid`;
  }

  if (keyword === "type" && prop?.type) {
    return `${label} must be a ${prop.type}`;
  }

  return `${label} is invalid`;
}

function parseInstanceLocation(location: string): string[] {
  if (!location || location === "#") return [];
  return location.replace(/^#\//, "").split("/").filter(Boolean);
}

function parseKeyword(location: string): string {
  if (!location) return "";
  const parts = location.split("/");
  return parts[parts.length - 1] || "";
}
