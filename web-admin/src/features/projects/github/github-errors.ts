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

  if (
    err.response.data?.message?.includes("worktree has additional contents")
  ) {
    return {
      message: "",
      notEmpty: true,
    };
  }

  return {
    message: err.response.data?.message,
    notEmpty: false,
  };
}

export function extractGithubDisconnectError(err: AxiosError<RpcStatus>) {
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

  const isLoginError =
    err.response.data?.message?.includes("refresh token is empty") ||
    err.response.data?.message?.includes(
      "refresh token passed is incorrect or expired.",
    );

  return {
    message: err.response.data?.message,
    isLoginError,
  };
}
