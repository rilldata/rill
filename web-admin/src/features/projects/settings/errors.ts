import type { RpcStatus } from "@rilldata/web-admin/client";
import type { AxiosError } from "axios";

export function parseUpdateProjectError(err: AxiosError<RpcStatus>) {
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
    err.response.data.message.includes(
      "a project with that name already exists",
    )
  ) {
    return {
      message: "",
      duplicateProject: true,
    };
  }

  return {
    message: err.response.data.message,
    duplicateProject: false,
  };
}
