import { setError, type SuperValidated } from "sveltekit-superforms";

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
  setError(form, SubmitErrorKey + ".message", error.message);
  setError(form, SubmitErrorKey + ".details", (error as any).details);
}

export function getSubmitError(errors: Record<string, any>): SubmitError {
  if (!errors[SubmitErrorKey]) return { message: "", details: "" };
  return {
    message: errors[SubmitErrorKey]?.message?.[0],
    details: errors[SubmitErrorKey]?.details?.[0],
  };
}
