import type { V1Message } from "@rilldata/web-common/runtime-client";
import { isHiddenAgentTool, MessageType, ToolName } from "../types";
import { createChartBlock, type ChartBlock } from "./chart/chart-block";
import { createTextMessage, type TextMessage } from "./text/text-message";
import {
  createThinkingBlock,
  hasDisplayableContent,
  type ThinkingBlock,
} from "./thinking/thinking-block";

/**
 * Planning indicator block - shown while waiting for the first AI response
 */
export type PlanningBlock = {
  type: "planning";
  id: string;
};

export type MessageBlock =
  | TextMessage
  | ThinkingBlock
  | ChartBlock
  | PlanningBlock;

// Re-export individual block types for convenience
export type { ChartBlock, TextMessage, ThinkingBlock };

/**
 * Transforms raw chat messages into a flat list of message blocks.
 * Handles grouping of thinking blocks, extracting charts, and adding planning indicators.
 */
export function transformToMessageBlocks(
  messages: V1Message[],
  isStreaming: boolean,
  isConversationLoading: boolean,
): MessageBlock[] {
  const blocks: MessageBlock[] = [];

  // 1. Build result message map for correlation with calls
  const resultMessagesByParentId = new Map(
    messages
      .filter(
        (msg) =>
          msg.type === MessageType.RESULT && msg.tool !== ToolName.ROUTER_AGENT,
      )
      .map((msg) => [msg.parentId, msg]),
  );

  // 2. Filter messages for processing
  // We keep router_agent results (assistant text) but filter out other tool results (handled via map)
  const displayMessages = messages.filter(
    (msg) =>
      msg.type !== MessageType.RESULT || msg.tool === ToolName.ROUTER_AGENT,
  );

  // 3. Group messages into blocks
  let currentThinkingMessages: V1Message[] | null = null;

  for (const msg of displayMessages) {
    if (msg.tool === ToolName.ROUTER_AGENT) {
      // Text message (user/assistant) - close any open thinking block
      finalizeThinkingBlock(currentThinkingMessages, messages, blocks);
      currentThinkingMessages = null;
      blocks.push(createTextMessage(msg, `text-${blocks.length}`));
    } else if (msg.type === MessageType.PROGRESS) {
      // Add to current thinking block or start a new one
      if (!currentThinkingMessages) currentThinkingMessages = [];
      currentThinkingMessages.push(msg);
    } else if (msg.type === MessageType.CALL) {
      // Tool calls are part of thinking block
      if (!currentThinkingMessages) currentThinkingMessages = [];
      currentThinkingMessages.push(msg);

      // Special handling for CREATE_CHART:
      // It ends the thinking block so the chart can be rendered at the top level
      if (msg.tool === ToolName.CREATE_CHART) {
        finalizeThinkingBlock(currentThinkingMessages, messages, blocks);
        currentThinkingMessages = null;

        // Add the chart block immediately after
        const chartBlock = createChartBlock(
          msg,
          resultMessagesByParentId.get(msg.id),
        );
        if (chartBlock) {
          blocks.push(chartBlock);
        }
      }
    }
  }

  // Close any remaining thinking block
  finalizeThinkingBlock(currentThinkingMessages, messages, blocks);

  // 4. Add Planning Indicator if needed
  if (
    shouldShowPlanningIndicator(messages, isStreaming, isConversationLoading)
  ) {
    blocks.push({
      type: "planning",
      id: "planning-indicator",
    });
  }

  return blocks;
}

/**
 * Finalizes and adds a thinking block to the blocks array if it has displayable content.
 */
function finalizeThinkingBlock(
  thinkingMessages: V1Message[] | null,
  allMessages: V1Message[],
  blocks: MessageBlock[],
): void {
  if (thinkingMessages && hasDisplayableContent(thinkingMessages)) {
    blocks.push(
      createThinkingBlock(
        thinkingMessages,
        allMessages,
        `thinking-${blocks.length}`,
      ),
    );
  }
}

/**
 * Determines if we should show the "Planning next moves..." indicator.
 */
function shouldShowPlanningIndicator(
  messages: V1Message[],
  isStreaming: boolean,
  isConversationLoading: boolean,
): boolean {
  // Must be actively loading/streaming
  if (!isStreaming && !isConversationLoading) return false;

  // Find the last user message
  const lastUserMessageIndex = messages.findLastIndex(
    (msg) => msg.role === "user",
  );

  if (lastUserMessageIndex === -1) return false;

  // Check if there are any visible AI messages after the last user message
  const hasAIResponseAfterUser = messages
    .slice(lastUserMessageIndex + 1)
    .some((msg) => {
      // 1. Progress messages are always visible responses
      if (msg.type === MessageType.PROGRESS) return true;

      // 2. Tool calls are responses if they are not hidden
      if (msg.type === MessageType.CALL) {
        return !isHiddenAgentTool(msg.tool);
      }

      // 3. Other assistant messages (Results/Text)
      if (msg.role === "assistant") {
        // Special case: Router Agent results are the main text response, so they count
        if (msg.tool === ToolName.ROUTER_AGENT) return true;

        // Other hidden tools should be ignored
        if (isHiddenAgentTool(msg.tool)) return false;

        // Default to true for any other visible assistant message
        return true;
      }

      return false;
    });

  return !hasAIResponseAfterUser;
}
