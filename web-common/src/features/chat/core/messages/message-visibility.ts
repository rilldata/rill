/**
 * Message Visibility
 *
 * Filters messages from the main chat block list. "Hidden" messages are still
 * rendered, just in different UI locations:
 *
 * - Internal tools (analyst_agent, etc.) → shown in thinking blocks
 * - Tool results → shown inside their parent tool call's collapsible UI
 * - Feedback messages → shown inline below the rated message
 *
 * Exception: `router_agent` results are the AI's text responses and ARE shown
 * as main chat blocks.
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageType, ToolName } from "../types";
import { isHiddenTool } from "./tools/tool-registry";

/**
 * Should this message be filtered from the main chat block list?
 */
export function shouldHideMessage(
  msg: V1Message,
  messageMap: Map<string, V1Message>,
): boolean {
  // Internal tools → rendered in thinking blocks (see tool-registry.ts)
  if (msg.tool !== ToolName.ROUTER_AGENT && isHiddenTool(msg.tool)) {
    return true;
  }

  // Tool results → rendered inside parent tool call's UI
  if (msg.type === MessageType.RESULT && msg.tool !== ToolName.ROUTER_AGENT) {
    return true;
  }

  // Feedback → rendered inline below the rated message
  if (isFeedbackRelated(msg, messageMap)) {
    return true;
  }

  return false;
}

// =============================================================================
// FEEDBACK VISIBILITY
// =============================================================================

/** Feedback messages are rendered by the feedback module, not as chat blocks. */
function isFeedbackRelated(
  msg: V1Message,
  messageMap: Map<string, V1Message>,
): boolean {
  return isFeedbackCall(msg) || isFeedbackResponse(msg, messageMap);
}

/** User initiated feedback (thumbs up/down button click). */
function isFeedbackCall(msg: V1Message): boolean {
  if (msg.tool !== ToolName.ROUTER_AGENT || msg.type !== MessageType.CALL) {
    return false;
  }
  try {
    const parsed = JSON.parse(msg.contentData || "");
    return !!parsed.user_feedback_args;
  } catch {
    return false;
  }
}

/** AI response to feedback (rendered inline by FeedbackState). */
function isFeedbackResponse(
  msg: V1Message,
  _messageMap: Map<string, V1Message>,
): boolean {
  if (msg.tool !== ToolName.ROUTER_AGENT || msg.type !== MessageType.RESULT) {
    return false;
  }

  try {
    const parsed = JSON.parse(msg.contentData || "");
    return parsed.agent === ToolName.USER_FEEDBACK;
  } catch {
    return false;
  }
}
