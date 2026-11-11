import { ConnectError, type Interceptor } from "@connectrpc/connect";
import { extractResourceName } from "./request-queue";

/** A runtime RPC, named the way query keys and error telemetry name it: `QueryService/metricsViewAggregation`. */
export interface AttributedRequest {
  api: string;
  // The metrics view or table the request was for, when its message names one
  resource?: string;
}

const MAX_REQUESTS_PER_ERROR = 5;

// Keyed by the error or abort reason that left a request; a WeakMap so the error is never kept alive by this
const requestsByError = new WeakMap<object, AttributedRequest[]>();

/**
 * Records, on every error or abort reason that leaves a runtime RPC, which RPC it belonged to.
 * A cancelled request that surfaces as an unhandled rejection is an AbortError with no stack and no hint of the request.
 * The request queue and fetch reject with the signal's reason, so that object is tagged when the signal aborts.
 * Connect's transport then rejects the call with `ConnectError.from` of whatever left the interceptors,
 * which for an AbortError is a new object that does not link to it;
 * so the interceptor converts the rejection to that ConnectError itself, tags both, and throws the ConnectError.
 * Register it as the outermost interceptor so the tag is in place before the queue rejects with the reason,
 * and so the error it throws is the one that leaves the transport; see `getAttributedRequests`.
 */
export function createRequestAttributionInterceptor(): Interceptor {
  return (next) => (req) => {
    const request: AttributedRequest = {
      api: `${shortServiceName(req.service.typeName)}/${localMethodName(req.method.name)}`,
    };
    const resource = extractResourceName(
      req.message as Record<string, unknown>,
    );
    if (resource) request.resource = resource;

    const onAbort = () => attributeErrorToRequest(req.signal.reason, request);
    req.signal.addEventListener("abort", onAbort, { once: true });
    return next(req).then(
      (res) => {
        // Connect aborts the signal after a successful call too; that reason reaches nobody
        req.signal.removeEventListener("abort", onAbort);
        return res;
      },
      (error: unknown) => {
        attributeErrorToRequest(error, request);
        // Returned as is when it already is a ConnectError
        const connectError = ConnectError.from(error);
        attributeErrorToRequest(connectError, request);
        throw connectError;
      },
    );
  };
}

/** The runtime RPCs an error or abort reason has been attributed to, or undefined if it did not leave a runtime RPC. */
export function getAttributedRequests(
  error: unknown,
): AttributedRequest[] | undefined {
  if (!error || typeof error !== "object") return undefined;
  return requestsByError.get(error);
}

function attributeErrorToRequest(error: unknown, request: AttributedRequest) {
  if (!error || typeof error !== "object") return;
  const requests = requestsByError.get(error);
  if (!requests) {
    requestsByError.set(error, [request]);
    return;
  }
  // The same request can see one reason twice (signal abort, then the rejection it causes)
  if (requests.includes(request) || requests.length >= MAX_REQUESTS_PER_ERROR) {
    return;
  }
  requests.push(request);
}

// "rill.runtime.v1.QueryService" => "QueryService"
function shortServiceName(typeName: string): string {
  return typeName.slice(typeName.lastIndexOf(".") + 1);
}

// "MetricsViewAggregation" => "metricsViewAggregation", the local name protobuf-es gives the method and query keys use
function localMethodName(name: string): string {
  return name.charAt(0).toLowerCase() + name.slice(1);
}
