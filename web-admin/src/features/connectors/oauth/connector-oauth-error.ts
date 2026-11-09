import { onQueryError } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { writable } from "svelte/store";

// Unanchored: ConnectError prefixes the message with the code, and some HTTP routes add their own prefix.
const signInRequiredRegex = /sign in to connector "([^"]+)" to view this data/;

/**
 * Returns the connector the user must sign in to when the error is the runtime's FailedPrecondition sign-in error, else null.
 */
export function connectorOAuthRequired(error: unknown): string | null {
  if (!error || typeof error !== "object") return null;
  const e = error as Record<string, unknown>;
  if (e.code !== 9 || typeof e.message !== "string") return null;
  return signInRequiredRegex.exec(e.message)?.[1] ?? null;
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
