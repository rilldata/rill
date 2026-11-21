import {
  extractErrorMessage,
  isErrorCode,
} from "@rilldata/web-common/lib/errors";
import { onQueryError } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { writable } from "svelte/store";

// Unanchored: some HTTP routes prefix the message.
const signInRequiredRegex = /sign in to connector "([^"]+)" to view this data/;

/**
 * Returns the connector the user must sign in to when the error is the runtime's FailedPrecondition sign-in error, else null.
 */
export function connectorOAuthRequired(error: unknown): string | null {
  if (!isErrorCode(error, 9)) return null;
  return signInRequiredRegex.exec(extractErrorMessage(error))?.[1] ?? null;
}

/** The connector the user must sign in to, set when a query fails because the user has not signed in. */
export const connectorSignInRequired = writable<string | null>(null);

/** Watches failed queries for sign-in errors. Returns a function that stops watching. */
export function watchConnectorSignInErrors(): () => void {
  return onQueryError((error) => {
    const connector = connectorOAuthRequired(error);
    if (connector) connectorSignInRequired.set(connector);
  });
}
