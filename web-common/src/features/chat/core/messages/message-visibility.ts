/**
 * Message Visibility
 *
 * Determines which messages should be hidden from the main chat message list.
 * "Hidden" here means not rendered as standalone blocks in the conversation flow.
 * These messages may still be visible through other UI elements (e.g., feedback
 * responses appear inline below the rated message, tool results are shown in
 * their parent tool call's collapsible section).
 *
 * Consolidates all visibility rules in one place, keeping block-transform
 * focused purely on transformation logic.
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageType, ToolName } from "../types";
import { isHiddenTool } from "./tools/tool-registry";

/**
 * Determines if a message should be hidden from the UI.
 *
 * Hidden messages include:
 * - Hidden tools (internal orchestration agents)
 * - Result messages (attached to parent calls via map)
 * - Feedback-related messages (calls, responses, internal processing)
 */
export function shouldHideMessage(
  msg: V1Message,
  messageMap: Map<string, V1Message>,
): boolean {
  // Hidden tools (from registry) - but not router_agent which produces text
  if (msg.tool !== ToolName.ROUTER_AGENT && isHiddenTool(msg.tool)) {
    return true;
  }

  // Result messages are attached to their parent calls, not rendered directly
  if (msg.type === MessageType.RESULT) {
    return true;
  }

  // Feedback-related messages
  if (isFeedbackRelated(msg, messageMap)) {
    return true;
  }

  return false;
}

// =============================================================================
// FEEDBACK VISIBILITY
// =============================================================================

/**
 * Check if a message is feedback-related and should be hidden.
 * Checks in message flow order: call → attribution → response.
 */
function isFeedbackRelated(
  msg: V1Message,
  messageMap: Map<string, V1Message>,
): boolean {
  return (
    isFeedbackCall(msg) ||
    isFeedbackAttribution(msg, messageMap) ||
    isFeedbackResponse(msg)
  );
}

/**
 * 1. Feedback call: router_agent message that initiates feedback submission.
 * Has user_feedback_args in its JSON content.
 */
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

/**
 * 2. Feedback attribution: intermediate messages from attribution prediction.
 * When negative feedback is submitted, the AI analyzes the conversation to
 * determine attribution (rill/project/user). These are children of the
 * user_feedback tool call but not the final result.
 */
function isFeedbackAttribution(
  msg: V1Message,
  messageMap: Map<string, V1Message>,
): boolean {
  if (!msg.parentId) return false;

  const parent = messageMap.get(msg.parentId);
  if (!parent) return false;

  return (
    parent.tool === ToolName.USER_FEEDBACK &&
    msg.tool !== ToolName.USER_FEEDBACK
  );
}

/**
 * 3. Feedback response: router_agent result containing the feedback acknowledgment.
 * Has agent: "user_feedback" in its result JSON.
 */
function isFeedbackResponse(msg: V1Message): boolean {
  if (msg.tool !== ToolName.ROUTER_AGENT || msg.type !== MessageType.RESULT) {
    return false;
  }
  try {
    const parsed = JSON.parse(msg.contentData || "");
    return parsed.agent === "user_feedback";
  } catch {
    return false;
  }
}
