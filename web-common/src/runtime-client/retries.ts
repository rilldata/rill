import type { HTTPError } from "./fetchWrapper";

export function shouldRetryConnectionError(
  failureCount: number,
  error: HTTPError,
) {
  if (error.message.includes("Failed to fetch")) {
    return failureCount < 3;
  }
  return false;
}
