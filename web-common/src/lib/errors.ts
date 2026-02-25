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
    typeof (error as any).response?.status === "number" &&
    typeof (error as any).response?.data?.message === "string"
  );
}
