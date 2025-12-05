/**
 * Message Block Transformation
 *
 * Transforms raw API messages (V1Message) into UI blocks (MessageBlock).
 *
 * Conceptual Model:
 * - `router_agent` messages → TEXT (the main conversation)
 * - `progress` messages → THINKING (collapsible reasoning UI)
 * - Tool calls → THINKING (inline) or BLOCK (top-level) based on registry
 * - `result` messages → Not rendered directly (attached to their parent calls)
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageType, ToolName } from "../types";
import { type ChartBlock } from "./chart/chart-block";
import { type FileDiffBlock } from "./file-diff/file-diff-block";
import {
  shouldShowPlanning,
  type PlanningBlock,
} from "./planning/planning-block";
import { createTextMessage, type TextMessage } from "./text/text-message";
import {
  createThinkingBlock,
  type ThinkingBlock,
} from "./thinking/thinking-block";
import { getToolConfig, type ToolConfig } from "./tool-registry";

// =============================================================================
// BLOCK TYPES
// =============================================================================

export type MessageBlock =
  | TextMessage
  | ThinkingBlock
  | ChartBlock
  | FileDiffBlock
  | PlanningBlock;

// Re-export individual block types for convenience
export type {
  ChartBlock,
  FileDiffBlock,
  PlanningBlock,
  TextMessage,
  ThinkingBlock,
};

// =============================================================================
// MESSAGE ROUTING
// =============================================================================

/**
 * Where a message should be rendered in the UI.
 */
export type MessageTarget =
  | { target: "text" }
  | { target: "thinking" }
  | { target: "block"; config: ToolConfig }
  | { target: "skip" };

/**
 * Determines where a message should be rendered.
 *
 * This centralizes all routing logic in one place:
 * - router_agent → text (main conversation)
 * - result messages → skip (attached to parent calls via map)
 * - progress messages → thinking
 * - tool calls → consult registry (inline/block/hidden)
 */
export function getMessageTarget(msg: V1Message): MessageTarget {
  // Router agent produces the main conversation text (user prompts, assistant responses)
  if (msg.tool === ToolName.ROUTER_AGENT) {
    return { target: "text" };
  }

  // Result messages are not directly rendered—they're attached to their parent calls
  if (msg.type === MessageType.RESULT) {
    return { target: "skip" };
  }

  // Progress messages always go to thinking block
  if (msg.type === MessageType.PROGRESS) {
    return { target: "thinking" };
  }

  // Tool calls: consult the registry
  if (msg.type === MessageType.CALL) {
    const config = getToolConfig(msg.tool);

    switch (config.renderMode) {
      case "block":
        return { target: "block", config };
      case "hidden":
        return { target: "skip" };
      case "inline":
      default:
        return { target: "thinking" };
    }
  }

  // Unknown message types are skipped
  return { target: "skip" };
}

// =============================================================================
// TRANSFORMATION
// =============================================================================

/**
 * Transforms raw chat messages into a list of UI blocks.
 */
export function transformToMessageBlocks(
  messages: V1Message[],
  isStreaming: boolean,
  isConversationLoading: boolean,
): MessageBlock[] {
  const blocks: MessageBlock[] = [];

  // Build result map for correlating tool calls with their results
  const resultMap = buildResultMessageMap(messages);

  // Accumulator for messages going into the current thinking block
  let thinkingMessages: V1Message[] = [];

  // Helper to flush the thinking block accumulator
  function flushThinking(isComplete: boolean): void {
    if (thinkingMessages.length > 0) {
      blocks.push(
        createThinkingBlock(
          thinkingMessages,
          resultMap,
          `thinking-${blocks.length}`,
          isComplete,
        ),
      );
      thinkingMessages = [];
    }
  }

  // Process each message
  for (const msg of messages) {
    const routing = getMessageTarget(msg);

    switch (routing.target) {
      case "text":
        // Text messages close any open thinking block
        flushThinking(true);
        blocks.push(createTextMessage(msg, `text-${blocks.length}`));
        break;

      case "thinking":
        // Accumulate in current thinking block
        thinkingMessages.push(msg);
        break;

      case "block": {
        // Block tools: add to thinking, flush it, then create the block
        thinkingMessages.push(msg);
        flushThinking(true);

        const block = routing.config.createBlock?.(msg, resultMap.get(msg.id));
        if (block) {
          blocks.push(block);
        }
        break;
      }

      case "skip":
        // Not rendered directly
        break;
    }
  }

  // Flush any remaining thinking messages
  const isRemainingComplete = !isStreaming && !isConversationLoading;
  flushThinking(isRemainingComplete);

  // Add planning indicator if needed
  if (shouldShowPlanning(messages, isStreaming, isConversationLoading)) {
    blocks.push({
      type: "planning",
      id: "planning-indicator",
    });
  }

  return blocks;
}

// =============================================================================
// HELPERS
// =============================================================================

/**
 * Build a map from tool call message IDs to their result messages.
 * Used to correlate tool calls with their results for display.
 */
function buildResultMessageMap(
  messages: V1Message[],
): Map<string | undefined, V1Message> {
  return new Map(
    messages
      .filter(
        (msg) =>
          msg.type === MessageType.RESULT && msg.tool !== ToolName.ROUTER_AGENT,
      )
      .map((msg) => [msg.parentId, msg]),
  );
}
