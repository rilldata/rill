import { SSEHttpError } from "@rilldata/web-common/runtime-client/sse";

/**
 * Format transport error into user-friendly message
 */
export function formatTransportError(error: Error): string {
  if (error.name === "AbortError") {
    return "Message sending was cancelled";
  }

  // Extract status code from SSEHttpError
  const status = error instanceof SSEHttpError ? error.status : null;

  // Authentication errors - suggest refresh to get new JWT
  if (status === 401 || status === 403) {
    return "Authentication failed. Please refresh the page and try again.";
  }

  // Bad request errors
  if (status === 400) {
    return "Invalid request. Please try again.";
  }

  // Server errors (5xx)
  if (status && status >= 500 && status < 600) {
    return "Server is temporarily unavailable. Please try sending your message again.";
  }

  // Rate limiting
  if (status === 429) {
    return "Too many requests. Please wait a moment before trying again.";
  }

  // Network/connection errors (fetch() throws TypeError for network failures)
  const lowerMessage = error.message?.toLowerCase() || "";
  const isNetworkError =
    (error.name === "TypeError" &&
      (lowerMessage.includes("fetch") ||
        lowerMessage.includes("network") ||
        lowerMessage.includes("load failed"))) ||
    (typeof navigator !== "undefined" && !navigator.onLine);

  if (isNetworkError) {
    return "Unable to connect to server. Please check your connection and try again.";
  }

  const isDeadlineExceeded =
    lowerMessage.includes("context deadline exceeded") ||
    lowerMessage.includes("agent timed out");
  if (isDeadlineExceeded) {
    return "Server took too long to respond. Please try again.";
  }

  // Fallback error message
  return "Failed to connect to server. Please try again or refresh the page.";
}
