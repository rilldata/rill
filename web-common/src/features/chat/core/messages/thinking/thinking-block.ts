import type { V1Message } from "@rilldata/web-common/runtime-client";
import { isHiddenAgentTool, MessageType, ToolName } from "../../types";

/**
 * Thinking block representation.
 * Contains progress messages and tool calls grouped together.
 */
export type ThinkingBlock = {
  type: "thinking";
  id: string;
  messages: V1Message[];
  isComplete: boolean;
  duration: number;
};

/**
 * Checks if a thinking block has any content worth displaying.
 */
export function hasDisplayableContent(messages: V1Message[]): boolean {
  return messages.some(
    (msg) =>
      msg.type === MessageType.PROGRESS ||
      (msg.type === MessageType.CALL && !isHiddenAgentTool(msg.tool)),
  );
}

/**
 * Determines if a thinking block is complete.
 * A thinking block is complete if:
 * 1. It ends with a CREATE_CHART tool call (which splits the block).
 * 2. There's a ROUTER_AGENT message after it.
 */
export function isThinkingBlockComplete(
  blockMessages: V1Message[],
  allMessages: V1Message[],
): boolean {
  if (blockMessages.length === 0) return false;

  // Find the last message in this thinking block
  const lastBlockMessage = blockMessages[blockMessages.length - 1];

  // If the block ends with a create_chart tool call, it is complete
  if (lastBlockMessage.tool === ToolName.CREATE_CHART) {
    return true;
  }

  const lastBlockIndex = allMessages.findIndex(
    (msg) => msg.id === lastBlockMessage.id,
  );

  if (lastBlockIndex === -1) return false;

  // Check if there's a ROUTER_AGENT message after this block
  for (let i = lastBlockIndex + 1; i < allMessages.length; i++) {
    const msg = allMessages[i];
    // If we find a ROUTER_AGENT message, the thinking block is complete
    if (msg.tool === ToolName.ROUTER_AGENT) {
      return true;
    }
  }

  // No ROUTER_AGENT message found after this block, so it's incomplete
  return false;
}

/**
 * Calculates the duration of a thinking block in seconds.
 * Uses the timestamps of the first and last messages.
 */
export function calculateThinkingDuration(messages: V1Message[]): number {
  if (messages.length === 0) return 0;

  const firstMessage = messages[0];
  const lastMessage = messages[messages.length - 1];

  if (!firstMessage.createdOn || !lastMessage.createdOn) return 0;

  const startTime = new Date(firstMessage.createdOn).getTime();
  const endTime = new Date(lastMessage.createdOn).getTime();

  const durationMs = endTime - startTime;
  return Math.max(0, Math.round(durationMs / 1000)); // Round to nearest second
}

/**
 * Creates a thinking block from messages.
 */
export function createThinkingBlock(
  messages: V1Message[],
  allMessages: V1Message[],
  id: string,
): ThinkingBlock {
  return {
    type: "thinking",
    id,
    messages,
    isComplete: isThinkingBlockComplete(messages, allMessages),
    duration: calculateThinkingDuration(messages),
  };
}
