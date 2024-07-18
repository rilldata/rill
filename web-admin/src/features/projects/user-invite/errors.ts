import type { RpcStatus } from "@rilldata/web-admin/client";
import type { AxiosError } from "axios";

export function parseError(error: AxiosError<RpcStatus>, email: string) {
  return `${email}: ${error.response?.data?.message ?? error.message}`;
}
