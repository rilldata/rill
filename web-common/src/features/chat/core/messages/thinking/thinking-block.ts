import type { V1Message } from "@rilldata/web-common/runtime-client";

/**
 * Thinking block representation.
 * Contains progress messages and tool calls grouped together.
 */
export type ThinkingBlock = {
  type: "thinking";
  id: string;
  messages: V1Message[];
  /** Map of tool call message IDs to their result messages */
  resultMessagesByParentId: Map<string | undefined, V1Message>;
  isComplete: boolean;
  duration: number;
};

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
 *
 * @param messages - The messages in this thinking block (already filtered to displayable)
 * @param resultMessagesByParentId - Map of tool call IDs to their result messages
 * @param id - Unique identifier for this block
 * @param isComplete - Whether this block is complete (no more messages expected)
 */
export function createThinkingBlock(
  messages: V1Message[],
  resultMessagesByParentId: Map<string | undefined, V1Message>,
  id: string,
  isComplete: boolean,
): ThinkingBlock {
  return {
    type: "thinking",
    id,
    messages,
    resultMessagesByParentId,
    isComplete,
    duration: calculateThinkingDuration(messages),
  };
}
