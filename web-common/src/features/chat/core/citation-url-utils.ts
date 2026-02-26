import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  type V1GetAIMessageResponse,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import type { Schema as MetricsResolverQuery } from "@rilldata/web-common/runtime-client/gen/resolvers/metrics/schema.ts";
import {
  MessageType,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";
import { error } from "@sveltejs/kit";

interface RuntimeInfo {
  host: string;
  instanceId: string;
  jwt?: { token: string } | undefined;
}

export async function fetchMessage(
  runtime: RuntimeInfo,
  conversationId: string,
  messageId: string,
) {
  try {
    const headers: Record<string, string> = {};
    if (runtime.jwt) {
      headers["Authorization"] = `Bearer ${runtime.jwt.token}`;
    }
    const resp = await fetch(
      `${runtime.host}/v1/instances/${runtime.instanceId}/ai/conversations/${conversationId}/messages/${messageId}`,
      { headers },
    );
    if (!resp.ok) {
      const data = await resp.json().catch(() => ({}));
      throw { response: { status: resp.status, data } };
    }
    const toolCallResp = (await resp.json()) as V1GetAIMessageResponse;

    // 200 response should always have a message.
    return toolCallResp.message!;
  } catch (e) {
    const apiError = e.response?.data?.message ?? "";
    switch (true) {
      case apiError.includes("failed to find the conversation"):
        throw error(
          404,
          new Error(
            "Failed to get conversation. Check if you have access to it.",
          ),
        );

      case apiError.includes("failed to find the call"):
        throw error(
          404,
          new Error(
            "Failed to get tool call for query. Check if you have access to it or if the call was deleted.",
          ),
        );
    }
    throw e;
  }
}

export function maybeGetMetricsResolverQueryFromMessage(message: V1Message) {
  if (
    message.tool !== ToolName.QUERY_METRICS_VIEW ||
    message.type !== MessageType.CALL
  ) {
    throw error(
      400,
      new Error(
        "Failed to get metrics resolver query, is not a metrics view query tool call.",
      ),
    );
  }

  const rawJson = message.content?.[0]?.toolCall?.input;
  if (!rawJson) {
    throw error(
      400,
      new Error(
        "Failed to get metrics resolver query, message has no content.",
      ),
    );
  }

  return rawJson as MetricsResolverQuery;
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
