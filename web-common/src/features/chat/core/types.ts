/**
 * Type constants and definitions for chat functionality
 *
 * Shared type constants that correspond to backend enums in runtime/ai/ai.go
 */
import type { RuntimeServiceCompleteBody } from "@rilldata/web-common/runtime-client";
import type { Readable } from "svelte/store";

// =============================================================================
// MESSAGE TYPE CONSTANTS
// =============================================================================

/**
 * Message types from the backend AI system
 * These correspond to MessageType enum in runtime/ai/ai.go
 */
export const MessageType = {
  CALL: "call",
  RESULT: "result",
  PROGRESS: "progress",
} as const;

/**
 * Message content types
 * These correspond to MessageContentType enum in runtime/ai/ai.go
 */
export const MessageContentType = {
  TEXT: "text",
  JSON: "json",
  ERROR: "error",
} as const;

/**
 * Tool names for agent and tool invocations
 */
export const ToolName = {
  ROUTER_AGENT: "router_agent",

  // Analyst Agent tools
  ANALYST_AGENT: "analyst_agent",
  LIST_METRICS_VIEWS: "list_metrics_views",
  GET_METRICS_VIEW: "get_metrics_view",
  QUERY_METRICS_VIEW_SUMMARY: "query_metrics_view_summary",
  QUERY_METRICS_VIEW: "query_metrics_view",
  CREATE_CHART: "create_chart",

  // Developer Agent tools
  DEVELOPER_AGENT: "developer_agent",
  DEVELOP_MODEL: "develop_model",
  DEVELOP_METRICS_VIEW: "develop_metrics_view",
  LIST_FILES: "list_files",
  SEARCH_FILES: "search_files",
  READ_FILE: "read_file",
  WRITE_FILE: "write_file",

  // Feedback agent
  FEEDBACK_AGENT: "feedback_agent",

  // Common tools
  NAVIGATE: "navigate",
} as const;

// =============================================================================
// CHAT CONFIG
// =============================================================================

export type ChatConfig = {
  agent: string;
  additionalContextStoreGetter: () => Readable<
    Partial<RuntimeServiceCompleteBody>
  >;
  emptyChatLabel: string;
  placeholder: string;
  minChatHeight: string;
};
