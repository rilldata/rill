/**
 * Classification config for the query hooks code generator.
 * Determines whether each RPC method produces a query or mutation hook,
 * or should be skipped entirely (streaming methods).
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
    // Streaming
    queryBatch: "skip",
  },
  RuntimeService: {
    // Streaming
    watchFiles: "skip",
    watchLogs: "skip",
    watchResources: "skip",
    completeStreaming: "skip",
    // Explicitly classified as queries (despite not matching Get/List prefix)
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
    const queryPrefixes = [
      "get",
      "list",
      "ping",
      "health",
      "instanceHealth",
    ];
    const lowerMethod = methodName.charAt(0).toLowerCase() + methodName.slice(1);
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
