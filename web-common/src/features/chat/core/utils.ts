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
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter";
import { DateTime, Interval } from "luxon";
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
          return (
            parsed.prompt ||
            parsed.response ||
            describeReportPrompt(parsed) ||
            rawContent
          );
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

/**
 * Scheduled AI reports start a conversation without a user prompt; the router call only carries the analyst agent's arguments.
 * Describe the report's scope (explore and time range) instead of showing the raw arguments.
 * Returns undefined if the message is not a report's opening call.
 */
function describeReportPrompt(routerArgs: {
  analyst_agent_args?: {
    is_report?: boolean;
    explore?: string;
    time_start?: string;
    time_end?: string;
    comparison_time_start?: string;
    comparison_time_end?: string;
  };
}): string | undefined {
  const args = routerArgs.analyst_agent_args;
  if (!args?.is_report) return undefined;

  const explore = args.explore ?? "";
  const timeRange =
    args.time_start && args.time_end
      ? formatReportTimeRange(args.time_start, args.time_end)
      : "";

  let prompt: string;
  if (explore && timeRange) {
    prompt = m.chat_report_prompt_explore_time_range({ explore, timeRange });
  } else if (explore) {
    prompt = m.chat_report_prompt_explore({ explore });
  } else if (timeRange) {
    prompt = m.chat_report_prompt_time_range({ timeRange });
  } else {
    prompt = m.chat_report_prompt();
  }

  if (args.comparison_time_start && args.comparison_time_end) {
    prompt = m.chat_report_prompt_comparison({
      prompt,
      comparisonTimeRange: formatReportTimeRange(
        args.comparison_time_start,
        args.comparison_time_end,
      ),
    });
  }
  return prompt;
}

// Report time ranges are resolved in the report's time zone (UTC by default) and are aligned to day boundaries there.
// Format them in UTC so the boundaries stay on whole days instead of picking up the viewer's offset.
function formatReportTimeRange(start: string, end: string): string {
  return prettyFormatTimeRange(
    Interval.fromDateTimes(
      DateTime.fromISO(start, { zone: "utc" }),
      DateTime.fromISO(end, { zone: "utc" }),
    ),
  );
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
