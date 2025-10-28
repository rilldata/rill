export function isEmpty(val: any) {
  return (
    val === undefined ||
    val === null ||
    val === "" ||
    (typeof val === "string" && val.trim() === "")
  );
}

export function normalizeErrors(
  err: any,
): string | string[] | null | undefined {
  if (!err) return undefined;
  if (Array.isArray(err)) return err;
  if (typeof err === "string") return err;
  if (err._errors && Array.isArray(err._errors)) return err._errors;
  return undefined;
}

import { humanReadableErrorMessage } from "../errors/errors";

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
