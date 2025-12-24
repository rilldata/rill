import { humanReadableErrorMessage } from "../errors/errors";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
import type { MultiStepFormSchema } from "./types";
import {
  findRadioEnumKey,
  getRadioEnumOptions,
  getRequiredFieldsByEnumValue,
} from "../../templates/schema-utils";
import { isStepMatch } from "./connector-schemas";

/**
 * Returns true for undefined, null, empty string, or whitespace-only string.
 * Useful for validating optional text inputs.
 */
export function isEmpty(val: any) {
  // Booleans are never empty (false is a valid value)
  if (typeof val === "boolean") return false;

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
export function normalizeErrors(
  err: any,
): string | string[] | null | undefined {
  if (!err) return undefined;
  if (Array.isArray(err)) return err;
  if (typeof err === "string") return err;
  if (err._errors && Array.isArray(err._errors)) return err._errors;
  return undefined;
}

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
  const props = connector?.configProperties ?? [];
  const hasDsn = props.some((p) => p.key === "dsn");
  const hasOthers = props.some((p) => p.key !== "dsn");
  return hasDsn && !hasOthers;
}

/**
 * Returns true when the active multi-step auth method has missing or invalid
 * required fields. Falls back to configured default/first auth method.
 */
export function isMultiStepConnectorDisabled(
  schema: MultiStepFormSchema | null,
  paramsFormValue: Record<string, unknown>,
  paramsFormErrors: Record<string, unknown>,
  currentStep: "connector" | "source" = "connector",
) {
  if (!schema || !paramsFormValue) return true;

  const authInfo = getRadioEnumOptions(schema);

  // Handle schemas without auth method radio selector (e.g., BigQuery, Athena)
  if (!authInfo) {
    const requiredFields = (schema.required ?? []).filter((fieldId) =>
      isStepMatch(schema, fieldId, currentStep),
    );
    if (!requiredFields.length) return false;

    return !requiredFields.every((fieldId) => {
      const value = paramsFormValue[fieldId];
      const errorsForField = paramsFormErrors[fieldId] as any;
      const hasErrors = Boolean(errorsForField?.length);
      return !isEmpty(value) && !hasErrors;
    });
  }

  const options = authInfo?.options ?? [];
  const authKey = authInfo?.key || findRadioEnumKey(schema);
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

  // When on source step and auth method isn't set (user skipped connector step),
  // validate all source step required fields regardless of auth method
  if (currentStep === "source" && !methodFromForm) {
    const allSourceRequired = new Set<string>();
    const requiredByMethod = getRequiredFieldsByEnumValue(schema, {
      step: currentStep,
    });

    // Collect all source step required fields across all auth methods
    for (const fields of Object.values(requiredByMethod)) {
      fields.forEach((field) => allSourceRequired.add(field));
    }

    const sourceRequiredFields = Array.from(allSourceRequired);
    if (!sourceRequiredFields.length) return false;

    return !sourceRequiredFields.every((fieldId) => {
      const value = paramsFormValue[fieldId];
      const errorsForField = paramsFormErrors[fieldId] as any;
      const hasErrors = Boolean(errorsForField?.length);
      return !isEmpty(value) && !hasErrors;
    });
  }

  const requiredByMethod = getRequiredFieldsByEnumValue(schema, {
    step: currentStep,
  });
  const requiredFields = requiredByMethod[method] ?? [];

  // If no required fields found for this step, button should be enabled
  if (!requiredFields.length) return false;

  // Check if all required fields are filled and have no errors
  return !requiredFields.every((fieldId) => {
    if (!isStepMatch(schema, fieldId, currentStep)) return true;
    const value = paramsFormValue[fieldId];
    const errorsForField = paramsFormErrors[fieldId] as any;
    const hasErrors = Boolean(errorsForField?.length);
    return !isEmpty(value) && !hasErrors;
  });
}
