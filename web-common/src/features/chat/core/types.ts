/**
 * Type constants and definitions for chat functionality
 *
 * Shared type constants that correspond to backend enums in runtime/ai/ai.go
 */

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

/**
 * Check if a tool call should be hidden from the UI
 */
export function isHiddenAgentTool(toolName: string | undefined): boolean {
  return !!toolName && HIDDEN_AGENT_TOOLS.includes(toolName);
}
