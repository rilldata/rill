/**
 * Display names for AI tools.
 * These match the openai/toolInvocation metadata from the backend tool specs.
 */

export interface ToolDisplayName {
  invoking: string; // Loading state - while tool is executing
  invoked: string; // Complete state - after tool has finished
}

/**
 * Tool display names mapping.
 * Sync these with the Meta annotations in runtime/ai/*_tool.go files.
 */
export const TOOL_DISPLAY_NAMES: Record<string, ToolDisplayName> = {
  router_agent: {
    invoking: "Routing prompt…",
    invoked: "Prompt completed",
  },
  analyst_agent: {
    invoking: "Analyzing…",
    invoked: "Completed analysis",
  },
  list_metrics_views: {
    invoking: "Listing metrics…",
    invoked: "Listed metrics",
  },
  get_metrics_view: {
    invoking: "Getting metrics definition…",
    invoked: "Found metrics definition",
  },
  query_metrics_view_summary: {
    invoking: "Querying metrics summary…",
    invoked: "Completed summary query",
  },
  query_metrics_view: {
    invoking: "Querying metrics…",
    invoked: "Completed metrics query",
  },
  create_chart: {
    invoking: "Creating chart…",
    invoked: "Finished creating chart",
  },
};

/**
 * Gets the display name for a tool based on its state.
 */
export function getToolDisplayName(
  toolName: string,
  isComplete: boolean,
): string {
  const displayName = TOOL_DISPLAY_NAMES[toolName];
  if (!displayName) {
    // Fallback to formatted tool name if not in mapping
    return formatToolName(toolName);
  }
  return isComplete ? displayName.invoked : displayName.invoking;
}

/**
 * Formats a tool name for display when no mapping exists.
 * Converts snake_case to Title Case.
 */
function formatToolName(toolName: string): string {
  return toolName
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
