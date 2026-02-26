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
