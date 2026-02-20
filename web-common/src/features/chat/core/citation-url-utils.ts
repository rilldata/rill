import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  getRuntimeServiceGetAIToolCallQueryKey,
  runtimeServiceGetAIToolCall,
} from "@rilldata/web-common/runtime-client";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";

export async function fetchQueryForCall(
  instanceId: string,
  conversationId: string,
  callId: string,
) {
  try {
    const toolCallResp = await queryClient.fetchQuery({
      queryKey: getRuntimeServiceGetAIToolCallQueryKey(
        instanceId,
        conversationId,
        callId,
      ),
      queryFn: () =>
        runtimeServiceGetAIToolCall(instanceId, conversationId, callId),
    });

    const query = toolCallResp.query as MetricsResolverQuery;
    return query;
  } catch (e) {
    const apiError = e.response?.data?.message ?? "";
    switch (true) {
      case apiError.includes("failed to find the conversation"):
        throw new Error(
          "Failed to get conversation. Check if you have access to it.",
        );

      case apiError.includes("failed to find the call"):
        throw new Error(
          "Failed to get tool call for query. Check if you have access to it or if the call was deleted.",
        );
    }
    throw e;
  }
}
const closingRoundBracketRegex = /\)$/;
const closingCurlyBracketRegex = /}$/;

// We sometimes see citation urls have an additional closing bracket. Here we try to parse by removing a trailing bracket.
// This is only relevant for older citation urls that had query as url param. New ones will reference the query metrics view call id.
export function getQueryFromUrl(url: URL) {
  // Get the JSON-encoded query parameters
  const queryJSON = url.searchParams.get("query");
  if (!queryJSON) {
    throw new Error("query parameter is required");
  }

  // Parse and validate the query with proper type safety
  let query: MetricsResolverQuery;
  try {
    // Try to parse the query as is
    query = JSON.parse(queryJSON) as MetricsResolverQuery;
  } catch (e) {
    // Next, try to parse the query with a trailing curly bracket removed
    try {
      query = JSON.parse(
        queryJSON.replace(closingCurlyBracketRegex, ""),
      ) as MetricsResolverQuery;
      return query;
    } catch {
      // no-op
    }

    // Finally, try to parse the query with a trailing round bracket removed
    try {
      query = JSON.parse(
        queryJSON.replace(closingRoundBracketRegex, ""),
      ) as MetricsResolverQuery;
      return query;
    } catch {
      // no-op
    }

    // If all else fails, throw the error from the original parse attempt
    throw new Error(`Invalid query: ${e.message}`);
  }

  return query;
}
