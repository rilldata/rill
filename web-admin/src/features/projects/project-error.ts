import {
  extractErrorMessage,
  extractErrorStatusCode,
} from "@rilldata/web-common/lib/errors";

/**
 * Statuses that mean the user can no longer reach the project:
 * the session expired (401), access was revoked (403), or the project was deleted or renamed (404).
 */
const ProjectUnreachableStatuses = new Set([401, 403, 404]);

export interface ProjectError {
  statusCode: number | undefined;
  message: string;
}

/**
 * Decides whether a failed `GetProject` query replaces the project page with an error page,
 * and what that page shows.
 *
 * `GetProject` polls in the background, and TanStack keeps the last successful response when a
 * refetch fails. While a cached project exists, only a 401, 403 or 404 replaces the page: the
 * cached project is misleading once the user can no longer reach it. Any other failure (a network
 * error, a timeout, a 5xx, rate limiting) is treated as transient and the cached project stays on
 * screen. Network errors additionally raise the "Showing cached data while we reconnect" banner
 * through `handleAdminServerNetworkError`.
 *
 * Without a cached project (the initial load failed), every error shows the error page.
 *
 * The status and message are read defensively: a network error or timeout has no `response`.
 */
export function resolveProjectError(
  error: unknown,
  hasCachedProject: boolean,
): ProjectError | undefined {
  if (!error) return undefined;

  const statusCode = extractErrorStatusCode(error);
  if (
    hasCachedProject &&
    (statusCode === undefined || !ProjectUnreachableStatuses.has(statusCode))
  ) {
    return undefined;
  }

  return { statusCode, message: extractErrorMessage(error) };
}
