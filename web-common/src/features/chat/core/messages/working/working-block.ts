/**
 * Working block representation.
 *
 * A visual indicator shown while the AI is still working.
 * Displays when streaming is in progress but the last block
 * is not a text block (since streaming text is its own indicator).
 */

import type { Block } from "../block-transform";

export type WorkingBlock = {
  type: "working";
  id: string;
};

/**
 * Determines if we should show the working indicator.
 *
 * Shows when streaming and:
 * - No blocks yet, OR
 * - Last block is NOT an assistant text (streaming text is its own indicator), AND
 * - Last block is NOT an incomplete thinking block (has its own shimmer)
 */
export function shouldShowWorking(
  blocks: Block[],
  isStreaming: boolean,
  isConversationLoading: boolean,
): boolean {
  if (!isStreaming && !isConversationLoading) return false;

  const lastBlock = blocks[blocks.length - 1];

  // No blocks yet - show indicator
  if (!lastBlock) return true;

  // Assistant text is streaming - text flow is the indicator
  if (lastBlock.type === "text" && lastBlock.message.role === "assistant") {
    return false;
  }

  // Incomplete thinking block - it has its own shimmer animation
  if (lastBlock.type === "thinking" && !lastBlock.isComplete) {
    return false;
  }

  return true;
}
