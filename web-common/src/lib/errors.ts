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
