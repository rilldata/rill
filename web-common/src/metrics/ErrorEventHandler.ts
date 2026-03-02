import { page } from "$app/stores";
import type { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import type {
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import {
  extractErrorMessage,
  extractErrorStatusCode,
} from "@rilldata/web-common/lib/errors";
import type { Query } from "@tanstack/query-core";
import { get } from "svelte/store";
import type {
  SourceConnectionType,
  SourceErrorCodes,
  SourceFileType,
} from "./service/SourceEventTypes";

export class ErrorEventHandler {
  public constructor(
    private readonly metricsService: MetricsService,
    private readonly commonUserMetrics: CommonUserFields,
    private readonly isDev: boolean,
    private readonly screenNameGetter: () => MetricsEventScreenName,
  ) {
    this.commonUserMetrics = commonUserMetrics;
  }

  public requestErrorEventHandler(error: unknown, query: Query) {
    const screenName = this.screenNameGetter();
    this.fireHTTPErrorBoundaryEvent(
      query.queryKey[0] as string,
      extractErrorStatusCode(error)?.toString() ?? "",
      extractErrorMessage(error),
      screenName,
      get(page).url.toString(),
    )?.catch(console.error);
  }

  public addJavascriptErrorListeners() {
    // Helper to detect errors originating from browser extensions
    const isExtensionError = (value: string | undefined) =>
      value?.includes("chrome-extension://") ||
      value?.includes("moz-extension://");

    const errorHandler = (errorEvt: ErrorEvent) => {
      // Ignore errors originating from browser extensions to avoid reporting issues outside our control
      if (isExtensionError(errorEvt.filename)) return;

      this.fireJavascriptErrorBoundaryEvent(
        errorEvt.error?.stack ?? "",
        errorEvt.message,
        this.screenNameGetter(),
        get(page).url.toString(),
      )?.catch(console.error);
    };

    const unhandledRejectionHandler = (
      rejectionEvent: PromiseRejectionEvent,
    ) => {
      let stack = "";
      let message = "";

      if (typeof rejectionEvent.reason === "string") {
        message = rejectionEvent.reason;
      } else if (rejectionEvent.reason instanceof Error) {
        stack = rejectionEvent.reason.stack ?? "";
        message = rejectionEvent.reason.message;
      } else {
        message = String.toString.apply(rejectionEvent.reason);
      }

      // Ignore errors originating from browser extensions
      if (isExtensionError(stack)) return;

      this.fireJavascriptErrorBoundaryEvent(
        stack,
        message,
        this.screenNameGetter(),
        get(page).url.toString(),
      )?.catch(console.error);
    };

    window.addEventListener("error", errorHandler);
    window.addEventListener("unhandledrejection", unhandledRejectionHandler);
    // return unsubscriber
    return () => {
      window.removeEventListener("error", errorHandler);
      window.removeEventListener(
        "unhandledrejection",
        unhandledRejectionHandler,
      );
    };
  }

  public fireSourceErrorEvent(
    space: MetricsEventSpace,
    screen_name: MetricsEventScreenName,
    error_code: SourceErrorCodes,
    connection_type: SourceConnectionType,
    file_type: SourceFileType,
    glob: boolean,
  ) {
    return this.metricsService.dispatch("sourceErrorEvent", [
      this.commonUserMetrics,
      space,
      screen_name,
      error_code,
      connection_type,
      file_type,
      glob,
    ]);
  }

  private fireHTTPErrorBoundaryEvent(
    api: string,
    status: string,
    message: string,
    screenName: MetricsEventScreenName,
    pageUrl: string,
  ) {
    if (this.isDev) return;

    return this.metricsService.dispatch("httpErrorEvent", [
      this.commonUserMetrics,
      screenName,
      api,
      status,
      message,
      pageUrl,
    ]);
  }

  private fireJavascriptErrorBoundaryEvent(
    stack: string,
    message: string,
    screenName: MetricsEventScreenName,
    pageUrl: string,
  ) {
    if (this.isDev) {
      console.log("javascriptErrorEvent", screenName, stack, message);
      return;
    }
    return this.metricsService.dispatch("javascriptErrorEvent", [
      this.commonUserMetrics,
      screenName,
      stack,
      message,
      pageUrl,
    ]);
  }
}
