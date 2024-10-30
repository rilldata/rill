import type { RpcStatus } from "@rilldata/web-admin/client";
import type { AxiosError } from "axios";

export function parseUpdateOrgError(err: AxiosError<RpcStatus>) {
  if (!err) {
    return {
      message: "",
    };
  }

  if (!err.response?.data?.message) {
    return {
      message: err.message,
    };
  }

  if (
    err.response.data.message.includes("an org with that name already exists")
  ) {
    return {
      message: "",
      duplicateOrg: true,
    };
  }

  return {
    message: err.response.data.message,
    duplicateOrg: false,
  };
}
