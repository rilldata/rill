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
  isCancelledRequestError,
} from "@rilldata/web-common/lib/errors";
import {
  type AttributedRequest,
  getAttributedRequests,
} from "@rilldata/web-common/runtime-client/v2/request-attribution";
import type { Query, QueryKey } from "@tanstack/query-core";
import { get } from "svelte/store";
import type {
  SourceConnectionType,
  SourceErrorCodes,
  SourceFileType,
} from "./service/SourceEventTypes";
import { categorizeSourceError } from "@rilldata/web-common/features/sources/errors/errors.ts";
import type { AddDataBehaviourEventFields } from "@rilldata/web-common/metrics/service/BehaviourEventTypes.ts";
import { getRecentResizeObserverCallbacks } from "./resize-observer-tracker";

// The telemetry pipeline drops `stack` today; keep it small in case it stops doing so
const MAX_STACK_CHARS = 3000;
const MAX_LISTED_ITEMS = 3;
const MAX_LISTED_SCRIPTS = 5;
const MAX_SCRIPT_URL_CHARS = 120;

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
      getApiNameFromQueryKey(query.queryKey),
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

      const described = describeJavascriptError(
        errorEvt.message,
        errorEvt.error,
        errorEvt.error?.stack ?? "",
      );
      this.fireJavascriptErrorBoundaryEvent(
        described.stack,
        described.message,
        this.screenNameGetter(),
        get(page).url.toString(),
        described.api,
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
        message = String(rejectionEvent.reason);
      }

      // Ignore errors originating from browser extensions
      if (isExtensionError(stack)) return;

      const described = describeJavascriptError(
        message,
        rejectionEvent.reason,
        stack,
      );
      this.fireJavascriptErrorBoundaryEvent(
        described.stack,
        described.message,
        this.screenNameGetter(),
        get(page).url.toString(),
        described.api,
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

  public fireAddDataErrorEvent(
    space: MetricsEventSpace,
    screen_name: MetricsEventScreenName,
    message: string,
    addDataFields: AddDataBehaviourEventFields,
  ) {
    const code = categorizeSourceError(message);
    return this.metricsService.dispatch("addDataErrorEvent", [
      this.commonUserMetrics,
      space,
      screen_name,
      code,
      addDataFields,
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
    api: string,
  ) {
    if (this.isDev) {
      console.log("javascriptErrorEvent", screenName, stack, message, api);
      return;
    }
    return this.metricsService.dispatch("javascriptErrorEvent", [
      this.commonUserMetrics,
      screenName,
      stack,
      message,
      pageUrl,
      api,
    ]);
  }
}

/**
 * Derives the `api` reported for a failed query from its query key.
 * ConnectRPC runtime queries use `[ServiceName, methodName, instanceId, request]` keys (see `runtime-client/v2/codegen/generator.ts`),
 * which are reported as `QueryService/metricsViewAggregation`.
 * Every other key, such as an admin REST path like `/v1/orgs/...` or a custom key like `metrics_sql`, reports its first element.
 * Dashboards group by this value, so it must stay short and stable: never include the instance id or the request.
 */
export function getApiNameFromQueryKey(queryKey: QueryKey): string {
  const [first, second] = queryKey;
  if (typeof first !== "string") return "";
  if (/^[A-Z]\w*Service$/.test(first) && typeof second === "string") {
    return `${first}/${second}`;
  }
  return first;
}

/**
 * Adds the signal needed to trace a JavaScript error: the runtime RPC an error or abort reason belonged to
 * (see `runtime-client/v2/request-attribution.ts`), the ResizeObserver callbacks that ran in the frame before a loop notice
 * (see `resize-observer-tracker.ts`), and the other script origins on the page behind a cross-origin "Script error.".
 * A short hint is appended to the message, which the telemetry pipeline keeps and groups by, and the detail goes in the stack.
 * The original message is kept in front so cancellations still match the pipeline's abort/cancel rule.
 */
function describeJavascriptError(
  message: string,
  error: unknown,
  stack: string,
): { message: string; stack: string; api: string } {
  let hint = "";
  let api = "";
  const details: string[] = [];

  const requests = getAttributedRequests(error);
  if (requests?.length) {
    api = requests[0].api;
    const kind = isCancelledRequestError(error)
      ? "aborted request"
      : "runtime request";
    hint = `${kind}: ${listed(unique(requests.map((r) => r.api)))}`;
    details.push(...requests.map(describeRequest));
  } else if (isCancelledRequestError(error)) {
    hint = "aborted request: no runtime request tagged";
  }

  if (message.includes("ResizeObserver loop")) {
    const callbacks = getRecentResizeObserverCallbacks();
    const targets = unique(callbacks.flatMap((callback) => callback.targets));
    hint = targets.length
      ? `last callbacks: ${listed(targets)}`
      : "no tracked callbacks in the last frame";
    const now = performance.now();
    details.push(
      ...callbacks.map(
        (callback) =>
          `${Math.round(callback.at - now)}ms: ${callback.targets.join(", ")}`,
      ),
    );
  }

  if (/^Script error\.?$/.test(message.trim())) {
    const scripts = thirdPartyScripts();
    const origins = unique(scripts.map((script) => script.hostname));
    hint = origins.length
      ? `cross-origin script; other script origins: ${listed(origins)}`
      : "cross-origin script; no third-party script tags";
    details.push(
      ...scripts
        .slice(0, MAX_LISTED_SCRIPTS)
        .map(
          (script) =>
            `script: ${script.src}` +
            (script.crossOrigin ? ` (crossorigin=${script.crossOrigin})` : ""),
        ),
    );
  }

  return {
    api,
    message: hint ? `${message} [${hint}]` : message,
    stack: [stack, ...details]
      .filter(Boolean)
      .join("\n")
      .slice(0, MAX_STACK_CHARS),
  };
}

function describeRequest(request: AttributedRequest): string {
  return (
    `request: ${request.api}` +
    (request.resource ? ` (${request.resource})` : "")
  );
}

/**
 * Script tags loaded from another origin: all the browser lets us know about a cross-origin "Script error.".
 * A src is reported as origin and path only: its credentials, query and fragment can carry tokens.
 */
function thirdPartyScripts(): Array<{
  src: string;
  hostname: string;
  crossOrigin: string | null;
}> {
  const scripts: Array<{
    src: string;
    hostname: string;
    crossOrigin: string | null;
  }> = [];
  for (const script of Array.from(document.scripts)) {
    if (!script.src) continue;
    try {
      const url = new URL(script.src, window.location.href);
      if (url.origin === window.location.origin) continue;
      scripts.push({
        src: (url.origin + url.pathname).slice(0, MAX_SCRIPT_URL_CHARS),
        hostname: url.hostname,
        crossOrigin: script.crossOrigin,
      });
    } catch {
      // An unparseable src says nothing about where the script came from
    }
  }
  return scripts;
}

function unique(items: string[]): string[] {
  return Array.from(new Set(items));
}

/** The first few items, then how many more there are */
function listed(items: string[]): string {
  const rest = items.length - MAX_LISTED_ITEMS;
  return (
    items.slice(0, MAX_LISTED_ITEMS).join(", ") + (rest > 0 ? ` +${rest}` : "")
  );
}
