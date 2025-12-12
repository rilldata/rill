import { derived } from "svelte/store";
import { createQuery } from "@tanstack/svelte-query";
import {
  getRuntimeServiceGetConversationQueryOptions,
  getRuntimeServiceListConversationsQueryOptions,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { MessageType } from "@rilldata/web-common/features/chat/core/types.ts";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";

/**
 * Looks at the last conversation and returns the metrics view used in the last message or tool call.
 */
export function getLastUsedMetricsViewNameStore() {
  const lastConversationQuery = createQuery(
    getLatestConversationQueryOptions(),
    queryClient,
  );

  return derived(lastConversationQuery, (latestConversation) => {
    if (!latestConversation.data?.messages) return null;

    for (const message of latestConversation.data.messages) {
      const metricsView = getMetricsViewInMessage(message);
      if (metricsView) return metricsView;
    }

    return null;
  });
}

/**
 * Returns the last updated conversation ID.
 */
function getLatestConversationQueryOptions() {
  const listConversationsQueryOptions = derived(runtime, ({ instanceId }) =>
    getRuntimeServiceListConversationsQueryOptions(instanceId, {
      // Filter to only show Rill client conversations, excluding MCP conversations
      userAgentPattern: "rill%",
    }),
  );
  const lastConversationId = derived(
    createQuery(listConversationsQueryOptions, queryClient),
    (conversationsResp) => {
      return conversationsResp?.data?.conversations?.[0]?.id;
    },
  );

  return derived(
    [lastConversationId, runtime],
    ([lastConversationId, { instanceId }]) => {
      return getRuntimeServiceGetConversationQueryOptions(
        instanceId,
        lastConversationId ?? "",
        {
          query: {
            enabled: !!lastConversationId,
          },
        },
      );
    },
  );
}

function getMetricsViewInMessage(message: V1Message) {
  if (message.type !== MessageType.CALL) return null;
  const content = message.content?.[0];
  return (content?.toolCall?.input?.metrics_view as string) ?? null;
}
