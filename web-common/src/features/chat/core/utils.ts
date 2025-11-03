/**
 * Shared utilities for chat functionality
 *
 * Common functions used across ConversationManager and Conversation classes to avoid duplication
 * and maintain consistency in error handling, ID generation, and cache management.
 */
import type { V1Message } from "@rilldata/web-common/runtime-client";

// =============================================================================
// ID GENERATION
// =============================================================================

export const NEW_CONVERSATION_ID = "new";

const OPTIMISTIC_MESSAGE_ID_PREFIX = "optimistic-message-";

export function getOptimisticMessageId(): string {
  return `${OPTIMISTIC_MESSAGE_ID_PREFIX}${Date.now()}`;
}

export function isOptimisticMessageId(id: string) {
  return id.startsWith(OPTIMISTIC_MESSAGE_ID_PREFIX);
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
// APP CONTEXT DETECTION
// =============================================================================

/**
 * Detect app context based on current route
 * Determines what resources/context the chat can see
 */
/*
// Commented out since V1AppContext is no longer used.
export function detectAppContext(page: Page): V1AppContext | null {
  const routeId = page.route.id;

  switch (routeId) {
    case "/[organization]/[project]/-/ai":
    case "/[organization]/[project]/-/ai/[conversationId]":
      return {
        contextType: V1AppContextType.APP_CONTEXT_TYPE_PROJECT_CHAT,
      };
    case "/[organization]/[project]/explore/[dashboard]":
      return {
        contextType: V1AppContextType.APP_CONTEXT_TYPE_EXPLORE_DASHBOARD,
        contextMetadata: {
          dashboard_name: page.params.dashboard,
        },
      };
    case "/(viz)/explore/[name]":
      return {
        contextType: V1AppContextType.APP_CONTEXT_TYPE_EXPLORE_DASHBOARD,
        contextMetadata: {
          dashboard_name: page.params.name,
        },
      };
    default:
      return null;
  }
}
*/

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

export function isUserMessage(message: V1Message) {
  return message.type && message.type === "user";
}
