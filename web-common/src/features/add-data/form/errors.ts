import {
  setError,
  type SuperForm,
  type SuperValidated,
} from "sveltekit-superforms";

/**
 * We use form error system to pass structured errors from connector/source creation.
 */

type SubmitError = {
  message: string;
  details: string;
};
const SubmitErrorKey = "submitError";

export function setSubmitError(
  form: SuperValidated<Record<string, unknown>>,
  error: Error,
) {
  setError(form, SubmitErrorKey + ".message", error.message, {
    removeFiles: false,
  });
  const details = (error as any).details;
  if (details) {
    setError(form, SubmitErrorKey + ".details", details, {
      removeFiles: false,
    });
  }
}

/**
 * Same as setSubmitError, for actions that run outside superform's submit flow.
 */
export function setSuperFormSubmitError(
  superForm: SuperForm<Record<string, unknown>>,
  error: Error,
) {
  const details = (error as any).details;
  superForm.errors.update((errors) => ({
    ...errors,
    [SubmitErrorKey]: {
      message: [error.message],
      ...(details ? { details: [details] } : {}),
    },
  }));
}

export function getSubmitError(errors: Record<string, any>): SubmitError {
  if (!errors[SubmitErrorKey]) return { message: "", details: "" };
  return {
    message: errors[SubmitErrorKey]?.message?.[0],
    details: errors[SubmitErrorKey]?.details?.[0],
  };
}
