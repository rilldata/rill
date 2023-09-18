import { goto } from "$app/navigation";
import { page } from "$app/stores";
import type { AxiosError } from "axios";
import { get } from "svelte/store";
import type { RpcStatus } from "../../client";
import { ADMIN_URL } from "../../client/http-client";
import { ErrorStoreState, errorStore } from "./error-store";

export function globalErrorCallback(error: AxiosError): void {
  const isProjectPage = get(page).route.id === "/[organization]/[project]";
  const isDashboardPage =
    get(page).route.id === "/[organization]/[project]/[dashboard]";

  // Special handling for some errors on the Project page
  if (isProjectPage) {
    // If "repository not found", ignore the error and show the page
    if (
      error.response.status === 400 &&
      (error.response.data as RpcStatus).message === "repository not found"
    ) {
      return;
    }
  }

  // Special handling for some errors on the Dashboard page
  if (isDashboardPage) {
    // If a dashboard wasn't found, let +page.svelte handle the error.
    // Because the project may be reconciling, in which case we want to show a loading spinner not a 404.
    if (
      error.response.status === 404 &&
      (error.response.data as RpcStatus).message === "not found"
    ) {
      return;
    }

    // When a JWT doesn't permit access to a metrics view, the metrics view APIs return 401s.
    // In this scenario, `GetCatalog` returns a 404. We ignore the 401s so we can show the 404.
    if (error.response.status === 401) {
      return;
    }
  }

  // If Unauthorized, redirect to login page
  if (error.response.status === 401) {
    goto(`${ADMIN_URL}/auth/login?redirect=${window.origin}`);
    return;
  }

  // Create a pretty message for the error page
  const errorStoreState = createErrorStoreStateFromAxiosError(error);

  errorStore.set(errorStoreState);
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
  } else if (msg === "project not found") {
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
