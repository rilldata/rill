/**
 * Shared utilities for chat functionality
 *
 * Common functions used across ConversationManager and Conversation classes to avoid duplication
 * and maintain consistency in error handling, ID generation, and cache management.
 */
import { useMetricsViewTimeRangeFromExplore } from "@rilldata/web-common/features/dashboards/selectors.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import { useExploreValidSpec } from "@rilldata/web-common/features/explores/selectors.ts";
import type { V1CompletionMessageContext } from "@rilldata/web-common/runtime-client";
import { derived, type Readable, readable } from "svelte/store";
import { useExploreState } from "../../dashboards/stores/dashboard-stores";
import { getTimeControlState } from "../../dashboards/time-controls/time-control-store";
import type { Page } from "@sveltejs/kit";

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
  if (!toolResult || !toolResult?.content || toolResult?.isError) return false;
  if (toolCall?.name !== "create_chart") return false;
  try {
    const parsed = JSON.parse(toolResult.content);
    return !!(parsed.chart_type && parsed.spec);
  } catch {
    return false;
  }
}

export function getDashboardContext(
  instanceId: string,
  page: Page,
): Readable<V1CompletionMessageContext | undefined> {
  const exploreName = page.params.name;
  if (!exploreName) return readable(undefined);

  return derived(
    [
      useExploreValidSpec(instanceId, exploreName),
      useMetricsViewTimeRangeFromExplore(instanceId, exploreName),
      useExploreState(exploreName),
    ],
    ([validSpecResp, timeRangeResp, exploreState]) => {
      const metricsViewSpec = validSpecResp.data?.metricsView ?? {};
      const exploreSpec = validSpecResp.data?.explore ?? {};

      const metricsViewName = exploreSpec.metricsView;
      if (!metricsViewName || !exploreState) return undefined;

      const timeControlState = getTimeControlState(
        metricsViewSpec,
        exploreSpec,
        timeRangeResp.data?.timeRangeSummary,
        exploreState,
      );
      let timeRange = "";
      if (
        timeControlState?.selectedTimeRange?.start &&
        timeControlState?.selectedTimeRange?.end
      ) {
        timeRange = `${timeControlState.selectedTimeRange.start.toISOString()} to ${timeControlState.selectedTimeRange.end.toISOString()}`;
      }

      let filters = "";
      if (exploreState.whereFilter?.cond?.exprs?.length > 0) {
        filters = convertExpressionToFilterParam(exploreState.whereFilter);
      }

      return <V1CompletionMessageContext>{
        metricsView: metricsViewName,
        explore: exploreName,
        timeRange,
        filters,
      };
    },
  );
}
