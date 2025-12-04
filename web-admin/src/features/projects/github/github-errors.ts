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
      message:
        "The subpath you specified already exists in this repo. Please use a different subpath or remove the existing folder before pushing.",
    };
  }

  if (
    err.response.data?.message?.includes("name already exists on this account")
  ) {
    return {
      message: "This repo already exists. Please choose a new repo name.",
    };
  }

  return {
    message:
      "Unable to complete push. Please check your repo and subpath settings and try again.",
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
