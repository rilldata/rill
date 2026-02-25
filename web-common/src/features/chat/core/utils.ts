/**
 * Shared utilities for chat functionality
 *
 * Common functions used across ConversationManager and Conversation classes to avoid duplication
 * and maintain consistency in ID generation and message content extraction.
 */
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import { type V1Message } from "@rilldata/web-common/runtime-client";
import {
  getRuntimeServiceGetConversationQueryOptions,
  getRuntimeServiceListConversationsQueryKey,
  getRuntimeServiceListConversationsQueryOptions,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import { MessageContentType, ToolName } from "./types";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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
export function getLatestConversationQueryOptions(client: RuntimeClient) {
  const listConversationsQueryOptions =
    getRuntimeServiceListConversationsQueryOptions(client, {
      // Filter to only show Rill client conversations, excluding MCP conversations
      userAgentPattern: "rill%",
    });
  const lastConversationId = derived(
    createQuery(listConversationsQueryOptions, queryClient),
    (conversationsResp) => {
      const conversations = conversationsResp?.data?.conversations?.filter(
        (c) => c.userAgent !== "rill/report",
      );
      return conversations?.[0]?.id;
    },
  );

  return derived([lastConversationId], ([id]) => {
    return getRuntimeServiceGetConversationQueryOptions(
      client,
      { conversationId: id ?? "" },
      {
        query: {
          enabled: !!id,
        },
      },
    );
  });
}
