import type { RpcStatus } from "@rilldata/web-admin/client";
import type { AxiosError } from "axios";

export function parseAccessRequestError(
  project: string,
  error: AxiosError<RpcStatus> | undefined,
) {
  if (!error) return "";
  if (!error.response?.data)
    return `Failed to approve access to ${project}: ${error.toString()}`;
  if (error.response.data.code === 5)
    return `Project access request not found. Perhaps the request was already accepted/denied.`;
  return `Failed to approve access to ${project}: ${error.response.data?.message}`;
}
