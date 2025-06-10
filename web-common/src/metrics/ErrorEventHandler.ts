import { page } from "$app/stores";
import type { RpcStatus } from "@rilldata/web-admin/client";
import type { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import type {
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { Query } from "@tanstack/query-core";
import type { AxiosError } from "axios";
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

  public requestErrorEventHandler(error: AxiosError, query: Query) {
    const screenName = this.screenNameGetter();
    if (!error.response) {
      this.fireHTTPErrorBoundaryEvent(
        query.queryKey[0] as string,
        error.status?.toString() ?? "",
        error.message ?? "unknown error",
        screenName,
        get(page).url.toString(),
      )?.catch(console.error);
      return;
    } else {
      this.fireHTTPErrorBoundaryEvent(
        query.queryKey[0] as string,
        error.response?.status?.toString() ?? error.status,
        (error.response?.data as RpcStatus)?.message ?? error.message,
        screenName,
        get(page).url.toString(),
      )?.catch(console.error);
    }
  }

  public addJavascriptErrorListeners() {
    const isExtensionError = (filename: string | undefined) =>
      filename?.startsWith("chrome-extension://") ||
      filename?.startsWith("moz-extension://");

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
