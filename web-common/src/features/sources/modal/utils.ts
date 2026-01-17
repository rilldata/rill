import { humanReadableErrorMessage } from "../errors/errors";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { ClickHouseConnectorType } from "./constants";
import type { MultiStepFormSchema } from "./types";
import {
  findRadioEnumKey,
  getRadioEnumOptions,
  getRequiredFieldsByEnumValue,
  getSchemaFieldMetaList,
  isStepMatch,
} from "../../templates/schema-utils";
import { getConnectorSchema } from "./connector-schemas";

/**
 * Returns true for undefined, null, empty string, or whitespace-only string.
 * Useful for validating optional text inputs.
 */
export function isEmpty(val: any) {
  return (
    val === undefined ||
    val === null ||
    val === "" ||
    (typeof val === "string" && val.trim() === "")
  );
}

/**
 * Normalizes a variety of error shapes into a string, string[], or undefined.
 * - If input is an array, returns it as-is.
 * - If input is a string, returns it.
 * - If input resembles a Zod `_errors` array, returns that.
 * - Otherwise returns undefined.
 */
/**
 * Converts unknown error inputs into a unified connector error shape.
 * - Prefers native Error.message when present
 * - Maps server error responses to human-readable messages via `humanReadableErrorMessage`
 * - Returns `details` with original message when it differs from the human-readable message
 */
export function normalizeConnectorError(
  connectorName: string,
  err: any,
): { message: string; details?: string } {
  let message: string;
  let details: string | undefined = undefined;

  if (err instanceof Error) {
    message = err.message;
  } else if (err?.message && err?.details) {
    message = err.message;
    details = err.details !== err.message ? err.details : undefined;
  } else if (err?.response?.data) {
    const originalMessage = err.response.data.message;
    const humanReadable = humanReadableErrorMessage(
      connectorName,
      err.response.data.code,
      originalMessage,
    );
    message = humanReadable;
    details = humanReadable !== originalMessage ? originalMessage : undefined;
  } else if (err?.message) {
    message = err.message;
  } else {
    message = "Unknown error";
  }

  return { message, details };
}

/**
 * Indicates whether a connector in "connector" mode exposes only a DSN field
 * (i.e., DSN exists and no other config properties are present).
 */
export function hasOnlyDsn(
  connector: V1ConnectorDriver | undefined,
  isConnectorForm: boolean,
): boolean {
  if (!isConnectorForm) return false;
  const schema = getConnectorSchema(connector?.name ?? "");
  if (schema) {
    const fields = getSchemaFieldMetaList(schema, { step: "connector" }).filter(
      (field) => !field.internal,
    );
    const hasDsn = fields.some((field) => field.key === "dsn");
    const hasOthers = fields.some((field) => field.key !== "dsn");
    return hasDsn && !hasOthers;
  }
  return false;
}

/**
 * Returns true when the active multi-step auth method has missing or invalid
 * required fields. Falls back to configured default/first auth method.
 */
export function isMultiStepConnectorDisabled(
  schema: MultiStepFormSchema | null,
  paramsFormValue: Record<string, unknown>,
  paramsFormErrors: Record<string, unknown>,
  step?: "connector" | "source" | string,
) {
  if (!schema) return true;

  // For source step, gate on required fields from the JSON schema.
  const currentStep = step || (paramsFormValue?.__step as string | undefined);
  if (currentStep === "source" || currentStep === "explorer") {
    const required = getRequiredFieldsForStep(
      schema,
      paramsFormValue,
      currentStep,
    );
    if (!required.length) return false;
    return !required.every((fieldId) => {
      if (!isStepMatch(schema, fieldId, currentStep)) return true;
      const value = paramsFormValue[fieldId];
      const errorsForField = paramsFormErrors[fieldId] as any;
      const hasErrors = Boolean(errorsForField?.length);
      return !isEmpty(value) && !hasErrors;
    });
  }

  const authInfo = getRadioEnumOptions(schema);
  const options = authInfo?.options ?? [];
  const authKey = authInfo?.key || findRadioEnumKey(schema);
  if (!authInfo || !options.length || !authKey) {
    const required = getRequiredFieldsForStep(
      schema,
      paramsFormValue,
      "connector",
    );
    if (!required.length) return false;
    return !required.every((fieldId) => {
      if (!isStepMatch(schema, fieldId, "connector")) return true;
      const value = paramsFormValue[fieldId];
      const errorsForField = paramsFormErrors[fieldId] as any;
      const hasErrors = Boolean(errorsForField?.length);
      return !isEmpty(value) && !hasErrors;
    });
  }
  const methodFromForm =
    authKey && paramsFormValue?.[authKey] != null
      ? String(paramsFormValue[authKey])
      : undefined;
  const hasValidFormSelection = options.some(
    (opt) => opt.value === methodFromForm,
  );
  const method =
    (hasValidFormSelection && methodFromForm) ||
    authInfo?.defaultValue ||
    options[0]?.value;

  if (!method) return true;

  // Selecting "public" should always enable the button for multi-step auth flows.
  if (method === "public") return false;

  const requiredByMethod = getRequiredFieldsByEnumValue(schema, {
    step: "connector",
  });
  const requiredFields = requiredByMethod[method] ?? [];
  if (!requiredFields.length) return true;

  return !requiredFields.every((fieldId) => {
    if (!isStepMatch(schema, fieldId, "connector")) return true;
    const value = paramsFormValue[fieldId];
    const errorsForField = paramsFormErrors[fieldId] as any;
    const hasErrors = Boolean(errorsForField?.length);
    return !isEmpty(value) && !hasErrors;
  });
}

function getRequiredFieldsForStep(
  schema: MultiStepFormSchema,
  values: Record<string, unknown>,
  step: "connector" | "source" | string,
) {
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

/**
 * Applies ClickHouse Cloud-specific default requirements for connector values.
 * - For ClickHouse Cloud: enforces `ssl: true`
 * - Otherwise returns values unchanged
 */
export function applyClickHouseCloudRequirements(
  connectorName: string | undefined,
  connectorType: ClickHouseConnectorType,
  values: Record<string, unknown>,
): Record<string, unknown> {
  // Only force SSL for ClickHouse Cloud when the user is using individual params.
  // DSN strings encapsulate their own protocol, so we should not inject `ssl` there.
  const isDsnBased = "dsn" in values;
  const shouldEnforceSSL =
    connectorName === "clickhouse" &&
    connectorType === "clickhouse-cloud" &&
    !isDsnBased;
  if (shouldEnforceSSL) {
    return { ...values, ssl: true } as Record<string, unknown>;
  }
  return values;
}
