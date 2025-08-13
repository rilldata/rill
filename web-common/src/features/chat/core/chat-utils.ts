/**
 * Shared utilities for chat functionality
 *
 * Common functions used across Chat and Conversation classes to avoid duplication
 * and maintain consistency in error handling, ID generation, and cache management.
 */
import {
  V1AppContextType,
  type V1AppContext,
} from "@rilldata/web-common/runtime-client";
import type { Page } from "@sveltejs/kit";

// =============================================================================
// ID GENERATION
// =============================================================================

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
// APP CONTEXT DETECTION
// =============================================================================

/**
 * Detect app context based on current route
 * Determines what resources/context the chat can see
 */
export function detectAppContext(page: Page): V1AppContext | null {
  const routeId = page.route.id;

  switch (routeId) {
    case "/[organization]/[project]/-/chat":
    case "/[organization]/[project]/-/chat/[conversationId]":
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
