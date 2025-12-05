import type { ConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
import { derived } from "svelte/store";
import { createQuery } from "@tanstack/svelte-query";
import {
  getRuntimeServiceGetConversationQueryOptions,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { MessageType } from "@rilldata/web-common/features/chat/core/types.ts";

/**
 * Looks at the last conversation and returns the metrics view used in the last message or tool call.
 */
export function getLastUsedMetricsViewNameStore(
  conversationManager: ConversationManager,
) {
  const lastConversationQuery = createQuery(
    getLatestConversationQueryOptions(conversationManager),
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
function getLatestConversationQueryOptions(
  conversationManager: ConversationManager,
) {
  const lastConversationId = derived(
    conversationManager.listConversationsQuery(),
    (conversationsResp) => {
      return conversationsResp?.data?.conversations?.[0]?.id;
    },
  );

  return derived(lastConversationId, (lastConversationId) => {
    return getRuntimeServiceGetConversationQueryOptions(
      conversationManager.instanceId,
      lastConversationId ?? "",
      {
        query: {
          enabled: !!lastConversationId,
        },
      },
    );
  });
}

function getMetricsViewInMessage(message: V1Message) {
  if (message.type !== MessageType.CALL) return null;
  const content = message.content?.[0];
  return content?.toolCall?.input?.metrics_view ?? null;
}
