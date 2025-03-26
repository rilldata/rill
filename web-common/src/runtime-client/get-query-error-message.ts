import type { RpcStatus } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import type { ErrorType } from "@rilldata/web-common/runtime-client/http-client";

export function getQueryErrorMessage(error: unknown) {
  if (typeof error === "string") return error;
  const rpcError = error as ErrorType<RpcStatus>;
  return (
    rpcError?.response?.data?.message ?? rpcError?.message ?? "Unknown error"
  );
}
