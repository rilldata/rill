import type { RpcStatus } from "@rilldata/web-admin/client";
import type { AxiosError } from "axios";

export function parseUpdateProjectError(err: AxiosError<RpcStatus> | null) {
  if (!err) return {};

  const message = err.response?.data?.message ?? err.message;

  return {
    duplicateProject: message?.includes("already exists"),
    message,
  };
}
