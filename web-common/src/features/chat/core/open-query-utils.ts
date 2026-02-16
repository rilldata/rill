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
