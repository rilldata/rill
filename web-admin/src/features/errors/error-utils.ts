import { page } from "$app/stores";
import { redirectToLogin } from "@rilldata/web-admin/client/redirect-utils";
import {
  isAdminServerQuery,
  isOrgUsageQuery,
} from "@rilldata/web-admin/client/utils";
import { redirectToLoginOrRequestAccess } from "@rilldata/web-admin/features/authentication/checkUserAccess";
import {
  isAlertPage,
  isMetricsExplorerPage,
  isProjectPage,
  isProjectRequestAccessPage,
  isPublicURLPage,
} from "@rilldata/web-admin/features/navigation/nav-utils";
import { errorEventHandler } from "@rilldata/web-common/metrics/initMetrics";
import {
  isGetResourceMetricsViewQuery,
  isRuntimeQuery,
} from "@rilldata/web-common/runtime-client/query-matcher";
import type { Query } from "@tanstack/query-core";
import type { QueryClient } from "@tanstack/svelte-query";
import type { AxiosError } from "axios";
import { get } from "svelte/store";
import type { RpcStatus } from "../../client";
import { getAdminServiceGetProjectQueryKey } from "../../client";
import { errorStore, type ErrorStoreState } from "./error-store";

export function createGlobalErrorCallback(queryClient: QueryClient) {
  return async (error: AxiosError, query: Query) => {
    errorEventHandler?.requestErrorEventHandler(error, query);

    const pageState = get(page);

    const onPublicURLPage = isPublicURLPage(pageState);
    if (onPublicURLPage) {
      // When a token is expired, show a specific error page
      if (
        error.response?.status === 401 &&
        (error.response.data as RpcStatus)?.message === "auth token is expired"
      ) {
        errorStore.set({
          statusCode: 401,
          header: "Oops! This link has expired",
          body: "It looks like this link is no longer active. Please reach out to the sender to request a new link.",
          fatal: true,
        });
        return;
      }

      // Let the Public URL page handle all other errors
      return;
    }

    // If an anonymous user hits a 403 error, redirect to the login page
    if (error.response?.status === 403) {
      const didRedirect = await redirectToLoginOrRequestAccess(pageState);
      if (didRedirect) return;
    }

    // If unauthorized to the admin server, redirect to login page
    if (isAdminServerQuery(query) && error.response?.status === 401) {
      redirectToLogin();
      return;
    }

    const onProjectPage = isProjectPage(pageState);

    // Special handling for some errors on the Project page
    if (onProjectPage) {
      if (error.response?.status === 400) {
        // If "repository not found", ignore the error and show the page
        if (
          (error.response.data as RpcStatus).message === "repository not found"
        ) {
          return;
        }

        // This error is the error:`driver.ErrNotFound` thrown while looking up an instance in the runtime.
        if (
          (error.response.data as RpcStatus).message === "driver: not found"
        ) {
          const [, org, proj] = pageState.url.pathname.split("/");
          void queryClient.resetQueries(
            getAdminServiceGetProjectQueryKey(org, proj),
          );
          return;
        }
      }

      // If the runtime throws a 401, it's likely due to a stale JWT that will soon be refreshed
      if (isRuntimeQuery(query) && error.response?.status === 401) {
        return;
      }
    }

    // Special handling for some errors on the Metrics Explorer page
    const onMetricsExplorerPage = isMetricsExplorerPage(pageState);
    if (onMetricsExplorerPage) {
      // Let the Metrics Explorer page handle errors for runtime queries.
      // Individual components (e.g. a specific line chart or leaderboard) should display a localised error message.
      if (isRuntimeQuery(query)) return;

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

    // Special handling for some errors on the Alerts page
    const onAlertPage = isAlertPage(pageState);
    if (onAlertPage) {
      // Don't block on a Metrics View 404
      if (
        isGetResourceMetricsViewQuery(query) &&
        error.response?.status === 404
      ) {
        return;
      }
    }

    // do not block on request access failures
    if (
      isProjectRequestAccessPage(pageState) &&
      error.response?.status !== 403
    ) {
      return;
    }

    // Handle case when usage metrics project is not unavailable for some reason.
    // We shouldn't block the user in this case.
    if (isOrgUsageQuery(query)) {
      return;
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
    detail: (error.response?.data as RpcStatus)?.message,
  };
}

export function createErrorPagePropsFromRoutingError(
  statusCode: number,
  errorMessage: string,
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
    detail: errorMessage,
  };
}
