export interface HTTPError {
  response: {
    status: number;
    data: {
      message: string;
    };
  };
  message: string;
  name: string;
  traceId?: string;
}

export function isHTTPError(error: unknown): error is HTTPError {
  return (
    typeof error === "object" &&
    error !== null &&
    typeof (error as Record<string, unknown>).response === "object" &&
    (error as Record<string, Record<string, unknown>>).response !== null &&
    typeof (error as { response: { status: unknown } }).response?.status ===
      "number" &&
    typeof (error as { response: { data: { message: unknown } } }).response
      ?.data?.message === "string"
  );
}

/**
 * Returns true if the error is a ConnectRPC NotFound error (code 5).
 * Works for any error shape: ConnectError, Axios/REST, or plain objects.
 */
export function isNotFoundError(error: unknown): boolean {
  if (!error || typeof error !== "object") return false;
  const e = error as Record<string, unknown>;
  // ConnectError: { code: 5 }
  if (e.code === 5) return true;
  // Axios/REST: { response: { status: 404 } } or { status: 404 }
  if (e.status === 404) return true;
  const resp = e.response;
  if (
    resp &&
    typeof resp === "object" &&
    (resp as Record<string, unknown>).status === 404
  )
    return true;
  return false;
}

/**
 * Extracts a human-readable error message from any error shape.
 * Handles ConnectError (rawMessage), HTTPError (response.data.message),
 * and plain Error (message).
 */
export function extractErrorMessage(error: unknown): string {
  if (!error || typeof error !== "object")
    return String(error ?? "Unknown error");
  const e = error as Record<string, unknown>;
  // ConnectRPC error
  if (typeof e.rawMessage === "string" && e.rawMessage) return e.rawMessage;
  // HTTPError (Axios/REST)
  const resp = e.response;
  if (resp && typeof resp === "object") {
    const data = (resp as Record<string, unknown>).data;
    if (data && typeof data === "object") {
      const msg = (data as Record<string, unknown>).message;
      if (typeof msg === "string" && msg) return msg;
    }
  }
  // Plain Error
  if (typeof e.message === "string" && e.message) return e.message;
  return "Unknown error";
}

/**
 * Extracts an HTTP status code from any error shape.
 * Handles ConnectError (gRPC code → HTTP mapping), HTTPError (response.status),
 * and plain objects (status).
 */
export function extractErrorStatusCode(error: unknown): number | undefined {
  if (!error || typeof error !== "object") return undefined;
  const e = error as Record<string, unknown>;
  // ConnectRPC error (gRPC codes 0–16)
  if (typeof e.code === "number" && e.code >= 0 && e.code <= 16) {
    return connectCodeToHTTPStatus(e.code);
  }
  // Direct status property
  if (typeof e.status === "number") return e.status;
  // HTTPError (Axios/REST)
  const resp = e.response;
  if (resp && typeof resp === "object") {
    const status = (resp as Record<string, unknown>).status;
    if (typeof status === "number") return status;
  }
  return undefined;
}

/**
 * Returns true if the error has the given ConnectRPC/gRPC code.
 * Also checks Axios-style response.data.code for backwards compatibility.
 */
export function isErrorCode(error: unknown, code: number): boolean {
  if (!error || typeof error !== "object") return false;
  const e = error as Record<string, unknown>;
  // ConnectRPC error
  if (e.code === code) return true;
  // Axios/REST: response.data.code
  const resp = e.response;
  if (resp && typeof resp === "object") {
    const data = (resp as Record<string, unknown>).data;
    if (data && typeof data === "object") {
      if ((data as Record<string, unknown>).code === code) return true;
    }
  }
  return false;
}

/**
 * Maps a ConnectRPC error code to an HTTP status code.
 * See https://connectrpc.com/docs/protocol#error-codes
 */
export function connectCodeToHTTPStatus(code: number): number {
  switch (code) {
    case 1: // Canceled
      return 408;
    case 2: // Unknown
      return 500;
    case 3: // InvalidArgument
      return 400;
    case 4: // DeadlineExceeded
      return 408;
    case 5: // NotFound
      return 404;
    case 6: // AlreadyExists
      return 409;
    case 7: // PermissionDenied
      return 403;
    case 8: // ResourceExhausted
      return 429;
    case 9: // FailedPrecondition
      return 412;
    case 10: // Aborted
      return 409;
    case 11: // OutOfRange
      return 400;
    case 12: // Unimplemented
      return 501;
    case 13: // Internal
      return 500;
    case 14: // Unavailable
      return 503;
    case 15: // DataLoss
      return 500;
    case 16: // Unauthenticated
      return 401;
    default:
      return 500;
  }
}
