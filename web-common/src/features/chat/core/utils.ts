/**
 * Shared utilities for chat functionality
 *
 * Common functions used across ConversationManager and Conversation classes to avoid duplication
 * and maintain consistency in ID generation and message content extraction.
 */
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  getRuntimeServiceGetConversationQueryOptions,
  getRuntimeServiceListConversationsQueryKey,
  getRuntimeServiceListConversationsQueryOptions,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import { MessageContentType, ToolName } from "./types";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store.ts";
import { derived } from "svelte/store";
import { createQuery } from "@tanstack/svelte-query";

// =============================================================================
// ID GENERATION
// =============================================================================

export const NEW_CONVERSATION_ID = "new";

const OPTIMISTIC_MESSAGE_ID_PREFIX = "optimistic-message-";

export function getOptimisticMessageId(): string {
  return `${OPTIMISTIC_MESSAGE_ID_PREFIX}${Date.now()}`;
}

// =============================================================================
// MESSAGE CONTENT EXTRACTION
// =============================================================================

/**
 * Extract text content from a message based on content type
 *
 * Handles all three content types (text, json, error) with special parsing
 * for router_agent JSON messages to extract prompt/response fields.
 */
export function extractMessageText(message: V1Message): string {
  const rawContent = message.contentData || "";

  switch (message.contentType) {
    case MessageContentType.JSON:
      // For router_agent, parse JSON and extract prompt/response field
      if (message.tool === ToolName.ROUTER_AGENT) {
        try {
          const parsed = JSON.parse(rawContent);
          return parsed.prompt || parsed.response || rawContent;
        } catch {
          return rawContent;
        }
      }

      // For non-router_agent JSON messages, return raw content
      return rawContent;

    case MessageContentType.TEXT:
      return rawContent;

    case MessageContentType.ERROR:
      return rawContent;

    default:
      return rawContent;
  }
}

export function invalidateConversationsList(instanceId: string) {
  const listConversationsKey = getRuntimeServiceListConversationsQueryKey(
    instanceId,
    {
      userAgentPattern: "rill%",
    },
  );
  return queryClient.invalidateQueries({ queryKey: listConversationsKey });
}

/**
 * Returns the last updated conversation ID.
 */
export function getLatestConversationQueryOptions() {
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
