import type { V1Message } from "@rilldata/web-common/runtime-client";
import { isHiddenAgentTool, MessageType, ToolName } from "../types";
import { createChartBlock, type ChartBlock } from "./chart/chart-block";
import { createTextBlock, type TextBlock } from "./text/text-block";
import {
  createPlanningIndicator,
  createThinkingBlock,
  hasDisplayableContent,
  type ThinkingBlock,
} from "./thinking/thinking-block";

export type MessageBlock = TextBlock | ThinkingBlock | ChartBlock;

// Re-export individual block types for convenience
export type { ChartBlock, TextBlock, ThinkingBlock };

/**
 * Transforms raw chat messages into a flat list of message blocks.
 * Handles grouping of thinking blocks, extracting charts, and adding planning indicators.
 */
export class MessageBlockTransformer {
  /**
   * Transforms raw chat messages into a flat list of message blocks.
   */
  public static transform(
    messages: V1Message[],
    resultMessagesByParentId: Map<string | undefined, V1Message>,
    isStreaming: boolean,
    isConversationLoading: boolean,
  ): MessageBlock[] {
    const blocks: MessageBlock[] = [];

    // 1. Filter messages for processing
    // We keep router_agent results (assistant text) but filter out other tool results (handled via map)
    const displayMessages = messages.filter(
      (msg) =>
        msg.type !== MessageType.RESULT || msg.tool === ToolName.ROUTER_AGENT,
    );

    // 2. Group messages into blocks
    let currentThinkingMessages: V1Message[] | null = null;

    const closeThinkingBlock = () => {
      if (
        currentThinkingMessages &&
        hasDisplayableContent(currentThinkingMessages)
      ) {
        blocks.push(
          createThinkingBlock(
            currentThinkingMessages,
            messages,
            `thinking-${blocks.length}`,
          ),
        );
      }
      currentThinkingMessages = null;
    };

    for (const msg of displayMessages) {
      if (msg.tool === ToolName.ROUTER_AGENT) {
        // Text message (user/assistant) - close any open thinking block
        closeThinkingBlock();
        blocks.push(createTextBlock(msg, `text-${blocks.length}`));
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
          closeThinkingBlock();

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
    closeThinkingBlock();

    // 3. Add Planning Indicator if needed
    if (
      MessageBlockTransformer.shouldShowPlanningIndicator(
        messages,
        isStreaming,
        isConversationLoading,
      )
    ) {
      blocks.push(createPlanningIndicator());
    }

    return blocks;
  }

  /**
   * Determines if we should show the "Planning next moves..." indicator.
   */
  private static shouldShowPlanningIndicator(
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

    // Check if there are any AI messages after the last user message
    const hasAIResponseAfterUser = messages
      .slice(lastUserMessageIndex + 1)
      .some(
        (msg) =>
          (msg.role === "assistant" && msg.tool !== ToolName.ROUTER_AGENT) || // Ignore router_agent calls which are user messages
          msg.type === MessageType.PROGRESS ||
          (msg.type === MessageType.CALL && !isHiddenAgentTool(msg.tool)),
      );

    return !hasAIResponseAfterUser;
  }
}
