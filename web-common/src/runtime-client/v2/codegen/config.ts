/**
 * Config for the query hooks code generator.
 * Determines whether each RPC method produces a query or mutation hook,
 * or should be skipped entirely (streaming methods), and whether its request
 * and response are proto messages or the legacy Orval JSON types.
 */

export type MethodClassification = "query" | "mutation" | "skip";

/**
 * Per-service overrides. Methods not listed here fall through to the
 * service-level default classifier.
 */
export const methodOverrides: Record<
  string,
  Record<string, MethodClassification>
> = {
  QueryService: {
    // Semantically write operations
    export: "mutation",
    exportReport: "mutation",
    query: "mutation",
  },
  RuntimeService: {
    // Instance-management: admin-plane only, not used by frontend
    createInstance: "skip",
    editInstance: "skip",
    deleteInstance: "skip",
    // Streaming
    watchFiles: "skip",
    watchLogs: "skip",
    watchResources: "skip",
    completeStreaming: "skip",
    // Explicitly classified as queries (despite not matching Get/List prefix)
    gitStatus: "query",
    gitDiff: "query",
    issueDevJWT: "query",
    analyzeConnectors: "query",
    analyzeVariables: "query",
    queryResolver: "query",
  },
  ConnectorService: {
    // All methods are queries (no overrides needed)
  },
};

/**
 * Default classifier for methods not in the override map.
 * RuntimeService uses prefix-based classification; other services default to query.
 */
export function classifyMethod(
  serviceName: string,
  methodName: string,
): MethodClassification {
  // Check overrides first
  const overrides = methodOverrides[serviceName];
  if (overrides && methodName in overrides) {
    return overrides[methodName];
  }

  // Service-specific defaults
  if (serviceName === "RuntimeService") {
    const queryPrefixes = ["get", "list", "ping", "health", "instanceHealth"];
    const lowerMethod =
      methodName.charAt(0).toLowerCase() + methodName.slice(1);
    if (queryPrefixes.some((p) => lowerMethod.startsWith(p))) {
      return "query";
    }
    return "mutation";
  }

  if (serviceName === "QueryService") {
    return "query";
  }

  if (serviceName === "ConnectorService") {
    return "query";
  }

  return "query";
}

/**
 * Methods migrated to proto messages, keyed by service.
 *
 * A migrated method takes a `PartialMessage<Request>` that is handed straight to
 * the ConnectRPC client and returns the `Response` message instance. Every other
 * method keeps the legacy JSON bridge: the request is parsed with `fromJson` and
 * the response is converted back with `toJson`, so callers keep working with the
 * Orval `V1*` JSON types.
 *
 * Migrating a method is a deliberate step, not a side effect of deleting a `V1*`
 * type from `index.schemas.ts`: add the method here in the same change that
 * updates its call sites to the proto messages.
 */
export const protoMessageMethods: Record<string, string[]> = {
  QueryService: [
    "columnCardinality",
    "columnDescriptiveStatistics",
    "columnNullCount",
    "columnNumericHistogram",
    "columnRollupInterval",
    "columnRugHistogram",
    "columnTimeGrain",
    "columnTimeRange",
    "columnTimeSeries",
    "columnTopK",
  ],
};

/** Whether a method's request and response are proto messages instead of Orval JSON types. */
export function usesProtoMessages(
  serviceName: string,
  methodName: string,
): boolean {
  return !!protoMessageMethods[serviceName]?.includes(methodName);
}
