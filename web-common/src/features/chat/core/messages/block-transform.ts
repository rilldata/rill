/**
 * Block Transformation
 *
 * Transforms raw API messages (V1Message) into UI blocks (Block).
 */

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageType, ToolName } from "../types";
import { type ChartBlock } from "./chart/chart-block";
import { type FileDiffBlock } from "./file-diff/file-diff-block";
import { shouldHideMessage } from "./message-visibility";
import { createTextBlock, type TextBlock } from "./text/text-block";
import {
  createThinkingBlock,
  type ThinkingBlock,
} from "./thinking/thinking-block";
import { getToolConfig, type ToolConfig } from "./tools/tool-registry";
import { shouldShowWorking, type WorkingBlock } from "./working/working-block";

// =============================================================================
// TYPES
// =============================================================================

export type Block =
  | TextBlock
  | ThinkingBlock
  | ChartBlock
  | FileDiffBlock
  | WorkingBlock;

export type {
  ChartBlock,
  FileDiffBlock,
  TextBlock,
  ThinkingBlock,
  WorkingBlock,
};

// =============================================================================
// TRANSFORMATION
// =============================================================================

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

  // Build message map for parent lookups (used by visibility checks)
  const messageMap = new Map(
    messages.filter((m) => m.id).map((m) => [m.id!, m]),
  );

  // Accumulator for messages going into the current thinking block
  let thinkingMessages: V1Message[] = [];

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
    const routing = getBlockRoute(msg, messageMap);

    switch (routing.route) {
      case "text":
        flushThinking(true);
        blocks.push(createTextBlock(msg));
        break;

      case "thinking":
        thinkingMessages.push(msg);
        break;

      case "block": {
        flushThinking(true);
        const block = routing.config.createBlock?.(msg, resultMap.get(msg.id));
        if (block) {
          blocks.push(block);
        }
        break;
      }

      case "skip":
        break;
    }
  }

  // Flush remaining thinking messages
  flushThinking(!isStreaming && !isConversationLoading);

  // Add working indicator if AI is still processing
  if (shouldShowWorking(blocks, isStreaming, isConversationLoading)) {
    blocks.push({ type: "working", id: "working-indicator" });
  }

  return blocks;
}

// =============================================================================
// ROUTING
// =============================================================================

type BlockRoute =
  | { route: "text" }
  | { route: "thinking" }
  | { route: "block"; config: ToolConfig }
  | { route: "skip" };

/**
 * Determines where a message should be routed for rendering.
 */
function getBlockRoute(
  msg: V1Message,
  messageMap: Map<string, V1Message>,
): BlockRoute {
  // Visibility check (handles hidden tools, results, feedback messages)
  if (shouldHideMessage(msg, messageMap)) {
    return { route: "skip" };
  }

  // Router agent → text (main conversation)
  if (msg.tool === ToolName.ROUTER_AGENT) {
    return { route: "text" };
  }

  // Progress → thinking
  if (msg.type === MessageType.PROGRESS) {
    return { route: "thinking" };
  }

  // Tool calls → consult registry for block vs inline
  if (msg.type === MessageType.CALL) {
    const config = getToolConfig(msg.tool);
    return config.renderMode === "block"
      ? { route: "block", config }
      : { route: "thinking" };
  }

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
