import type { RpcStatus } from "@rilldata/web-admin/client";
import type { AxiosError } from "axios";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

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
      message: m.github_error_subpath_exists(),
    };
  }

  if (
    err.response.data?.message?.includes("name already exists on this account")
  ) {
    return {
      message: m.github_error_repo_exists(),
    };
  }

  return {
    message: m.github_error_push_failed(),
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
