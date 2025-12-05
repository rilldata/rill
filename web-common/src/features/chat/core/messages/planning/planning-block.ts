/**
 * Planning block representation.
 *
 * Shown while waiting for the AI to start responding.
 * This is a "branch point" - the next block could be either:
 * - ThinkingBlock (if AI uses tools/reasoning)
 * - TextMessage (if AI responds directly)
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { getBlockRoute } from "../block-transform";

export type PlanningBlock = {
  type: "planning";
  id: string;
};

/**
 * Determines if we should show the "Planning next moves..." indicator.
 * Shows when streaming/loading but no visible AI response has arrived yet.
 */
export function shouldShowPlanning(
  messages: V1Message[],
  isStreaming: boolean,
  isConversationLoading: boolean,
): boolean {
  if (!isStreaming && !isConversationLoading) return false;

  // Find the last user message
  const lastUserIndex = messages.findLastIndex((msg) => msg.role === "user");
  if (lastUserIndex === -1) return false;

  // Check if there's any visible AI content after the last user message
  const hasVisibleResponse = messages.slice(lastUserIndex + 1).some((msg) => {
    const routing = getBlockRoute(msg);
    return routing.route !== "skip";
  });

  return !hasVisibleResponse;
}
