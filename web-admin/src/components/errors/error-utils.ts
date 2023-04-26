import { goto } from "$app/navigation";
import type { AxiosError } from "axios";
import type { RpcStatus } from "../../client";
import { ADMIN_URL } from "../../client/http-client";
import { ErrorStoreState, errorStore, isErrorStoreEmpty } from "./error-store";

export function globalErrorCallback(error: AxiosError): void {
  // If Unauthorized, redirect to login page
  if (error.response.status === 401) {
    goto(`${ADMIN_URL}/auth/login?redirect=${window.origin}`);
    return;
  }

  // Create a pretty message for the error page
  const errorStoreState = createErrorStoreStateFromAxiosError(error);

  // Sometimes there are multiple errors in one page load. When one of them is a 404, we
  // prioritize it by overwriting whatever error was already in the store.
  if (isErrorStoreEmpty || errorStoreState.statusCode === 404) {
    errorStore.set(errorStoreState);
  }
}

function createErrorStoreStateFromAxiosError(
  error: AxiosError
): ErrorStoreState {
  const status = error.response.status;
  const msg = (error.response.data as RpcStatus).message;

  // Specifically handle some errors
  if (status === 403) {
    return {
      statusCode: error.response.status,
      header: "Access denied",
      body: "You don't have access to this page. Please check that you have the correct permissions.",
    };
  } else if (msg === "org not found") {
    return {
      statusCode: error.response.status,
      header: "Organization not found",
      body: "The organization you requested could not be found. Please check that you have provided a valid organization name.",
    };
  } else if (msg === "proj not found") {
    return {
      statusCode: error.response.status,
      header: "Project not found",
      body: "The project you requested could not be found. Please check that you have provided a valid project name.",
    };
  }

  // Fallback for all other errors
  return {
    statusCode: error.response.status,
    header: "Sorry, unexpected error!",
    body: "Try refreshing the page, and reach out to us if that doesn't fix the error.",
  };
}

export function createErrorPagePropsFromRoutingError(
  statusCode: number
): ErrorStoreState {
  if (statusCode === 404) {
    return {
      statusCode: 404,
      header: "Sorry, we can't find this page!",
      body: "The page you're looking for might have been removed, had its name changed, or is temporarily unavailable.",
    };
  }

  // Not expecting any other errors, but here's a fallback just in case.
  return {
    statusCode: statusCode,
    header: "Sorry, unexpected error!",
    body: "Try refreshing the page, and reach out to us if that doesn't fix the error.",
  };
}
