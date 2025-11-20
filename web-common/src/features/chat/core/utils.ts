/**
 * Shared utilities for chat functionality
 *
 * Common functions used across ConversationManager and Conversation classes to avoid duplication
 * and maintain consistency in error handling, ID generation, and cache management.
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
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
// ERROR HANDLING
// =============================================================================

/**
 * Standardized error message formatting for chat functionality
 */
export function formatChatError(error: unknown): string {
  if (error instanceof Error) {
    if (
      error.message.includes("instance") ||
      error.message.includes("API client")
    ) {
      return error.message;
    } else if (
      error.message.includes("Network Error") ||
      error.message.includes("fetch")
    ) {
      return "Could not connect to the server. Please check your connection.";
    } else {
      return "An unexpected error occurred. Please try again.";
    }
  }

  return typeof error === "string"
    ? error
    : "An unexpected error occurred. Please try again.";
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

// =============================================================================
// CHART UTILITIES
// =============================================================================

// Helper to check if a tool result contains chart data
export function isChartToolResult(toolResult: any, toolCall: any): boolean {
  if (toolResult?.isError || toolCall?.name !== "create_chart") return false;
  try {
    // Check if input is already an object or needs parsing
    const parsed =
      typeof toolCall?.input === "string"
        ? JSON.parse(toolCall.input)
        : toolCall?.input;
    return !!(parsed?.chart_type && parsed?.spec);
  } catch {
    return false;
  }
}

// Helper to parse chart data from tool result
export function parseChartData(toolCall: any) {
  try {
    // Check if input is already an object or needs parsing
    const parsed =
      typeof toolCall?.input === "string"
        ? JSON.parse(toolCall.input)
        : toolCall?.input;

    return {
      chartType: parsed.chart_type,
      chartSpec: parsed.spec,
    };
  } catch (error) {
    console.error("Failed to parse chart data:", error);
    return null;
  }
}
