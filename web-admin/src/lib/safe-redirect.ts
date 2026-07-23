export type RedirectPolicy = {
  allowedOrigins?: readonly string[];
  allowLoopback?: boolean;
};

const LOOPBACK_HOSTS = new Set(["localhost", "127.0.0.1", "[::1]"]);

/**
 * Resolves redirects against the current page and fails closed. A destination
 * must be same-origin, explicitly configured, or (when requested) an HTTP
 * loopback callback used by a local client.
 */
export function getSafeRedirectTarget(
  redirect: string | null | undefined,
  currentUrl: string,
  policy: RedirectPolicy = {},
): string | null {
  if (!redirect?.trim()) return null;

  try {
    const current = new URL(currentUrl);
    const target = new URL(redirect, current);

    if (
      (target.protocol !== "http:" && target.protocol !== "https:") ||
      target.username ||
      target.password
    ) {
      return null;
    }

    if (target.origin === current.origin) return target.toString();

    const allowedOrigins = new Set(
      (policy.allowedOrigins ?? []).flatMap((origin) => {
        try {
          return [new URL(origin).origin];
        } catch {
          return [];
        }
      }),
    );
    if (allowedOrigins.has(target.origin)) return target.toString();

    if (
      policy.allowLoopback &&
      target.protocol === "http:" &&
      LOOPBACK_HOSTS.has(target.hostname)
    ) {
      return target.toString();
    }
  } catch {
    // Malformed input is an unapproved destination, not a navigation request.
  }

  return null;
}

/** Exact origins in this build-time allowlist may receive browser redirects. */
export function getConfiguredRedirectOrigins(
  value: string | undefined = import.meta.env
    .RILL_UI_PUBLIC_ALLOWED_REDIRECT_ORIGINS as string | undefined,
): string[] {
  return (value ?? "")
    .split(",")
    .map((origin) => origin.trim())
    .filter(Boolean);
}
