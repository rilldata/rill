import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "../core/types";

// =============================================================================
// BACKEND TYPES (mirror runtime/ai/memory.go)
// =============================================================================

/** One applied memory operation, as returned by update_memory and extract_memories. */
export interface MemoryOp {
  op: "add" | "update" | "delete" | "noop";
  memory_id?: string;
  category?: string;
  content?: string;
  previous_category?: string;
  previous_content?: string;
  previous_status?: string;
  reason?: string;
}

interface MemoryUpdateResultData {
  ops?: MemoryOp[];
}

// =============================================================================
// BLOCK TYPE
// =============================================================================

/**
 * A "Memory updated" notice: the memory operations applied during a turn,
 * either explicitly through update_memory or in the background through extract_memories.
 */
export type MemoryUpdateBlock = {
  type: "memory-update";
  id: string;
  message: V1Message;
  resultMessage: V1Message;
  /** Applied operations (noops are dropped) */
  ops: MemoryOp[];
};

/**
 * Creates a memory update block from an update_memory or extract_memories tool call.
 * Returns null while the result is pending, when the call failed, or when nothing changed.
 */
export function createMemoryUpdateBlock(
  message: V1Message,
  resultMessage: V1Message | undefined,
): MemoryUpdateBlock | null {
  if (!resultMessage) return null;
  if (resultMessage.contentType === MessageContentType.ERROR) return null;

  const ops = parseAppliedOps(resultMessage.contentData);
  if (ops.length === 0) return null;

  return {
    type: "memory-update",
    id: `memory-update-${message.id}`,
    message,
    resultMessage,
    ops,
  };
}

function parseAppliedOps(contentData: string | undefined): MemoryOp[] {
  if (!contentData) return [];
  try {
    const data = JSON.parse(contentData) as MemoryUpdateResultData;
    return (data.ops ?? []).filter((op) => op.op !== "noop" && !!op.memory_id);
  } catch {
    return [];
  }
}
