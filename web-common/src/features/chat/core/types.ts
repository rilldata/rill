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
  ANALYST_AGENT: "analyst_agent",
  DEVELOPER_AGENT: "developer_agent",
  CREATE_CHART: "create_chart",
  QUERY_METRICS_VIEW: "query_metrics_view",
  LIST_FILES: "list_files",
  READ_FILE: "read_file",
  WRITE_FILE: "write_file",
  DEVELOP_MODEL: "develop_model",
  DEVELOP_METRICS_VIEW: "develop_metrics_view",
} as const;

// =============================================================================
// TOOL FILTERING
// =============================================================================

/**
 * High-level agent tools that should not be rendered in the UI
 * These are internal orchestration agents, not user-facing tools
 */
const HIDDEN_AGENT_TOOLS: readonly string[] = [
  ToolName.ROUTER_AGENT,
  ToolName.ANALYST_AGENT,
  ToolName.DEVELOPER_AGENT,
];

export type ChatConfig = {
  agent: string;
  additionalContextStoreGetter: () => Readable<
    Partial<RuntimeServiceCompleteBody>
  >;
  emptyChatLabel: string;
  placeholder: string;
  enableMention: boolean; // TODO: should be a list of allowed mentions in the future
};

/**
 * Check if a tool call should be hidden from the UI
 */
export function isHiddenAgentTool(toolName: string | undefined): boolean {
  return !!toolName && HIDDEN_AGENT_TOOLS.includes(toolName);
}
