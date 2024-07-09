import type { RpcStatus } from "@rilldata/web-admin/client";
import type { AxiosError } from "axios";

export function parseAccessRequestError(
  error: AxiosError<RpcStatus> | undefined,
) {
  if (!error) return "";
  if (!error.response?.data) return error.toString();
  if (error.response.data.code === 5)
    return `Project access request not found. Perhaps the request was already accepted/denied.`;
  return error.response.data.message;
}
