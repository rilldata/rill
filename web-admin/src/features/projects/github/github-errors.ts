import type { RpcStatus } from "@rilldata/web-admin/client";
import type { AxiosError } from "axios";

export function extractGithubConnectError(err: AxiosError<RpcStatus>) {
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

  if (err.response.data.message.includes("worktree has additional contents")) {
    return {
      message: "",
      notEmpty: true,
    };
  }

  return {
    message: err.response.data.message,
    notEmpty: false,
  };
}
