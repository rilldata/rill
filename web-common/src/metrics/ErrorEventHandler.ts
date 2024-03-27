import type { RpcStatus } from "@rilldata/web-admin/client";
import type { MetricsService } from "@rilldata/web-common/metrics/service/MetricsService";
import type {
  CommonUserFields,
  MetricsEventScreenName,
  MetricsEventSpace,
} from "@rilldata/web-common/metrics/service/MetricsTypes";
import type { Query } from "@tanstack/query-core";
import type { AxiosError } from "axios";
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
        error.status ?? "",
        error.message ?? "unknown error",
        screenName,
      );
      return;
    } else {
      this.fireHTTPErrorBoundaryEvent(
        query.queryKey[0] as string,
        error.response?.status + "" ?? error.status,
        (error.response?.data as RpcStatus)?.message ?? error.message,
        screenName,
      );
    }
  }

  public addJavascriptErrorListeners() {
    const errorHandler = (errorEvt: ErrorEvent) => {
      this.fireJavascriptErrorBoundaryEvent(
        errorEvt.error?.stack ?? "",
        errorEvt.message,
        this.screenNameGetter(),
      );
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
      );
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
  ) {
    if (this.isDev) return;

    return this.metricsService.dispatch("httpErrorEvent", [
      this.commonUserMetrics,
      screenName,
      api,
      status,
      message,
    ]);
  }

  private fireJavascriptErrorBoundaryEvent(
    stack: string,
    message: string,
    screenName: MetricsEventScreenName,
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
    ]);
  }
}
