/**
 * Block Transformation
 *
 * Transforms raw API messages (V1Message) into UI blocks (Block).
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageType, ToolName } from "../types";
import { type ChartBlock } from "./chart/chart-block";
import { type FileDiffBlock } from "./file-diff/file-diff-block";
import { createTextBlock, type TextBlock } from "./text/text-block";
import {
  createThinkingBlock,
  type ThinkingBlock,
} from "./thinking/thinking-block";
import { getToolConfig, type ToolConfig } from "./tools/tool-registry";
import { shouldShowWorking, type WorkingBlock } from "./working/working-block";
import type { SimpleToolCall } from "@rilldata/web-common/features/chat/core/messages/simple-tool-call/simple-tool-call.ts";

// =============================================================================
// TYPES & TRANSFORMATION
// =============================================================================

export type Block =
  | TextBlock
  | ThinkingBlock
  | ChartBlock
  | FileDiffBlock
  | WorkingBlock
  | SimpleToolCall;

// Re-export individual block types for convenience
export type {
  ChartBlock,
  FileDiffBlock,
  TextBlock,
  ThinkingBlock,
  WorkingBlock,
  SimpleToolCall,
};

/**
 * Transforms raw chat messages into a list of UI blocks.
 */
export function transformToBlocks(
  messages: V1Message[],
  isStreaming: boolean,
  isConversationLoading: boolean,
): Block[] {
  const blocks: Block[] = [];

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
    const routing = getBlockRoute(msg);

    switch (routing.route) {
      case "text":
        // Text blocks close any open thinking block
        flushThinking(true);
        blocks.push(createTextBlock(msg));
        break;

      case "thinking":
        // Accumulate in current thinking block
        thinkingMessages.push(msg);
        break;

      case "block": {
        // Block tools render their own header, so flush thinking first
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

  // Add working indicator if AI is still processing
  if (shouldShowWorking(blocks, isStreaming, isConversationLoading)) {
    blocks.push({
      type: "working",
      id: "working-indicator",
    });
  }

  return blocks;
}

// =============================================================================
// ROUTING & HELPERS
// =============================================================================

/**
 * Where a message should be routed for rendering.
 */
type BlockRoute =
  | { route: "text" }
  | { route: "thinking" }
  | { route: "block"; config: ToolConfig }
  | { route: "skip" };

/**
 * Determines where a message should be routed.
 *
 * Routing rules:
 * - router_agent → text (main conversation)
 * - result messages → skip (attached to parent calls via map)
 * - progress messages → thinking
 * - tool calls → consult registry (inline/block/hidden)
 */
function getBlockRoute(msg: V1Message): BlockRoute {
  // Router agent produces the main conversation text
  if (msg.tool === ToolName.ROUTER_AGENT) {
    return { route: "text" };
  }

  // Result messages are attached to their parent calls, not rendered directly
  if (msg.type === MessageType.RESULT) {
    return { route: "skip" };
  }

  // Progress messages always go to thinking block
  if (msg.type === MessageType.PROGRESS) {
    return { route: "thinking" };
  }

  // Tool calls: consult the registry
  if (msg.type === MessageType.CALL) {
    const config = getToolConfig(msg.tool);

    switch (config.renderMode) {
      case "block":
        return { route: "block", config };
      case "hidden":
        return { route: "skip" };
      case "inline":
      default:
        return { route: "thinking" };
    }
  }

  // Unknown message types are skipped
  return { route: "skip" };
}

/**
 * Build a map from tool call message IDs to their result messages.
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
