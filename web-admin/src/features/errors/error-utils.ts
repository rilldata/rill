import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { isAdminServerQuery } from "@rilldata/web-admin/client/utils";
import {
  isDashboardPage,
  isProjectPage,
} from "@rilldata/web-admin/features/navigation/nav-utils";
import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
import { isRuntimeQuery } from "@rilldata/web-common/runtime-client/is-runtime-query";
import type { Query } from "@tanstack/query-core";
import type { QueryClient } from "@tanstack/svelte-query";
import type { AxiosError } from "axios";
import { get } from "svelte/store";
import type { RpcStatus } from "../../client";
import { getAdminServiceGetProjectQueryKey } from "../../client";
import { ADMIN_URL } from "../../client/http-client";
import { ErrorStoreState, errorStore } from "./error-store";

export function createGlobalErrorCallback(queryClient: QueryClient) {
  return (error: AxiosError, query: Query) => {
    errorEventHandler?.requestErrorEventHandler(error, query);

    // If unauthorized to the admin server, redirect to login page
    if (isAdminServerQuery(query) && error.response?.status === 401) {
      goto(
        `${ADMIN_URL}/auth/login?redirect=${window.location.origin}${window.location.pathname}`,
      );
      return;
    }

    // Special handling for some errors on the Project page
    const onProjectPage = isProjectPage(get(page));
    if (onProjectPage && error.response?.status === 400) {
      // If "repository not found", ignore the error and show the page
      if (
        (error.response.data as RpcStatus).message === "repository not found"
      ) {
        return;
      }
      // This error is the error:`driver.ErrNotFound` thrown while looking up an instance in the runtime.
      if ((error.response.data as RpcStatus).message === "driver: not found") {
        const [, org, proj] = get(page).url.pathname.split("/");
        queryClient.resetQueries(getAdminServiceGetProjectQueryKey(org, proj));
        return;
      }
    }

    // Special handling for some errors on the Dashboard page
    const onDashboardPage = isDashboardPage(get(page));
    if (onDashboardPage) {
      // Let the Dashboard page handle errors for runtime queries.
      // Individual components (e.g. a specific line chart or leaderboard) should display a localised error message.
      // NOTE: let's start with 400 errors, but we may want to include 500-level errors too.
      if (
        isRuntimeQuery(query) &&
        (error.response?.status === 400 || error.response?.status === 429)
      ) {
        return;
      }

      // If a dashboard wasn't found, let +page.svelte handle the error.
      // Because the project may be reconciling, in which case we want to show a loading spinner not a 404.
      if (
        error.response?.status === 404 &&
        (error.response.data as RpcStatus).message === "not found"
      ) {
        return;
      }

      // When a JWT doesn't permit access to a metrics view, the metrics view APIs return 401s.
      // In this scenario, `GetCatalog` returns a 404. We ignore the 401s so we can show the 404.
      if (error.response?.status === 401) {
        return;
      }
    }

    // Create a pretty message for the error page
    const errorStoreState = createErrorStoreStateFromAxiosError(error);

    // Show the error page
    errorStore.set(errorStoreState);
  };
}

function createErrorStoreStateFromAxiosError(
  error: AxiosError,
): ErrorStoreState {
  // Handle network errors
  if (error.message === "Network Error") {
    return {
      statusCode: null,
      header: "Network Error",
      body: "It seems we're having trouble reaching our servers. Check your connection or try again later.",
    };
  }

  // Handle some application errors
  const status = error.response?.status;
  const msg = (error?.response?.data as RpcStatus | undefined)?.message;
  if (status === 403) {
    return {
      statusCode: error.response.status,
      header: "Access denied",
      body: "You don't have access to this page. Please check that you have the correct permissions.",
    };
  } else if (msg === "org not found") {
    return {
      statusCode: error.response?.status,
      header: "Organization not found",
      body: "The organization you requested could not be found. Please check that you have provided a valid organization name.",
    };
  } else if (msg === "project not found") {
    return {
      statusCode: error.response?.status,
      header: "Project not found",
      body: "The project you requested could not be found. Please check that you have provided a valid project name.",
    };
  } else if (status === 400 && msg === "driver: not found") {
    return {
      statusCode: error.response?.status,
      header: "Project deployment not found",
      body: "This is potentially a temporary state if the project has just been reset.",
    };
  }

  // Fallback for all other errors (including 5xx errors)
  return {
    statusCode: error.response?.status,
    header: "Sorry, unexpected error!",
    body: "Try refreshing the page, and reach out to us if that doesn't fix the error.",
  };
}

export function createErrorPagePropsFromRoutingError(
  statusCode: number,
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
