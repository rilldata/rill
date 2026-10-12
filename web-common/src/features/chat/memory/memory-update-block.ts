import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "../core/types";

// =============================================================================
// BACKEND TYPES (mirror runtime/ai/memory_update.go)
// =============================================================================

/** The outcome of an update_memory call: the applied operation, or a noop with the reason it was rejected. */
export interface MemoryUpdateResult {
  op: "add" | "update" | "delete" | "noop";
  memory_id?: string;
  content?: string;
  reason?: string;
}

// =============================================================================
// BLOCK TYPE
// =============================================================================

/** A "Memory updated" notice: the memory operation the AI applied during a turn. */
export type MemoryUpdateBlock = {
  type: "memory-update";
  id: string;
  message: V1Message;
  resultMessage: V1Message;
  /** The applied operation (never a noop) */
  result: MemoryUpdateResult;
};

/**
 * Creates a memory update block from an update_memory tool call.
 * Returns null while the result is pending, when the call failed, or when nothing changed.
 */
export function createMemoryUpdateBlock(
  message: V1Message,
  resultMessage: V1Message | undefined,
): MemoryUpdateBlock | null {
  if (!resultMessage) return null;
  if (resultMessage.contentType === MessageContentType.ERROR) return null;

  const result = parseAppliedResult(resultMessage.contentData);
  if (!result) return null;

  return {
    type: "memory-update",
    id: `memory-update-${message.id}`,
    message,
    resultMessage,
    result,
  };
}

function parseAppliedResult(
  contentData: string | undefined,
): MemoryUpdateResult | null {
  if (!contentData) return null;
  try {
    const result = JSON.parse(contentData) as MemoryUpdateResult;
    if (result.op === "noop" || !result.memory_id) return null;
    return result;
  } catch {
    return null;
  }
}
