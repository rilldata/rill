/**
 * Utility functions for thinking block logic
 */

import type { V1Message } from "../../../../runtime-client";

/**
 * Determines if a thinking block is complete.
 * A thinking block is complete only when there's a ROUTER_AGENT message after it.
 * This indicates the AI has finished thinking and provided a response.
 */
export function isThinkingBlockComplete(
  blockMessages: V1Message[],
  allMessages: V1Message[],
): boolean {
  if (blockMessages.length === 0) return false;

  // Find the last message in this thinking block
  const lastBlockMessage = blockMessages[blockMessages.length - 1];
  const lastBlockIndex = allMessages.findIndex(
    (msg) => msg.id === lastBlockMessage.id,
  );

  if (lastBlockIndex === -1) return false;

  // Check if there's a ROUTER_AGENT message after this block
  for (let i = lastBlockIndex + 1; i < allMessages.length; i++) {
    const msg = allMessages[i];
    // If we find a ROUTER_AGENT message, the thinking block is complete
    if (msg.tool === "router_agent") {
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
 * Formats duration in seconds to a human-readable string.
 * Examples: "1 second", "5 seconds", "1 minute 30 seconds"
 */
export function formatThinkingDuration(seconds: number): string {
  if (seconds < 1) return "less than a second";
  if (seconds === 1) return "1 second";
  if (seconds < 60) return `${seconds} seconds`;

  const minutes = Math.floor(seconds / 60);
  const remainingSeconds = seconds % 60;

  if (remainingSeconds === 0) {
    return minutes === 1 ? "1 minute" : `${minutes} minutes`;
  }

  return `${minutes} ${minutes === 1 ? "minute" : "minutes"} ${remainingSeconds} ${remainingSeconds === 1 ? "second" : "seconds"}`;
}
