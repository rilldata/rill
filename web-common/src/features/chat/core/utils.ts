/**
 * Shared utilities for chat functionality
 *
 * Common functions used across ConversationManager and Conversation classes to avoid duplication
 * and maintain consistency in ID generation and message content extraction.
 */
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  getRuntimeServiceListConversationsQueryKey,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import { MessageContentType, ToolName } from "./types";

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
