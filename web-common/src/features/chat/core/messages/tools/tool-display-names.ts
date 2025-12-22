/**
 * Utilities for getting display names from AI tools.
 * Display names are fetched from the ListTools API and stored in tool.meta.
 */

import type { V1Tool } from "@rilldata/web-common/runtime-client";

// Meta keys for tool invocation display names (matches backend conventions)
const META_KEY_INVOKING = "openai/toolInvocation/invoking";
const META_KEY_INVOKED = "openai/toolInvocation/invoked";

/**
 * Gets the display name for a tool based on its completion state.
 * Uses the tool's meta field from the ListTools API response.
 */
export function getToolDisplayName(
  toolName: string,
  isComplete: boolean,
  tools: V1Tool[] | undefined,
): string {
  const tool = tools?.find((t) => t.name === toolName);

  if (tool?.meta) {
    const metaKey = isComplete ? META_KEY_INVOKED : META_KEY_INVOKING;
    const displayName = tool.meta[metaKey];
    if (typeof displayName === "string") {
      return displayName;
    }
  }

  // Fallback to formatted tool name if no meta or tool not found
  return formatToolName(toolName);
}

/**
 * Formats a tool name for display when no meta exists.
 * Converts snake_case to Title Case.
 */
function formatToolName(toolName: string): string {
  return toolName
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
