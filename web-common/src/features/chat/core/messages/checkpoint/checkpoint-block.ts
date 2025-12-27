import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "../../types";

// =============================================================================
// BACKEND TYPES (mirror runtime/ai tool definitions)
// =============================================================================

/** Arguments for the checkpoint tool call */
interface CheckpointCallData {
  // Empty - checkpoint takes no arguments
}

/** Result from the checkpoint tool */
interface CheckpointResultData {
  sha: string;
  message: string;
  had_changes: boolean;
}

// =============================================================================
// BLOCK TYPE
// =============================================================================

/**
 * Checkpoint block representation.
 * Contains commit information from a checkpoint tool call.
 */
export type CheckpointBlock = {
  type: "checkpoint";
  id: string;
  message: V1Message;
  resultMessage: V1Message;
  commitSha: string;
  commitMessage: string;
  hadChanges: boolean;
};

/**
 * Creates a checkpoint block from a checkpoint tool call message.
 * Returns null if the data is invalid or the result indicates an error.
 */
export function createCheckpointBlock(
  message: V1Message,
  resultMessage: V1Message | undefined,
): CheckpointBlock | null {
  if (!resultMessage) return null;
  if (resultMessage.contentType === MessageContentType.ERROR) return null;

  try {
    const resultData: CheckpointResultData = JSON.parse(
      resultMessage.contentData || "{}",
    );

    const commitSha = resultData.sha || "";
    if (!commitSha) return null;

    return {
      type: "checkpoint",
      id: `checkpoint-${message.id}`,
      message,
      resultMessage,
      commitSha,
      commitMessage: resultData.message || "",
      hadChanges: resultData.had_changes || false,
    };
  } catch {
    return null;
  }
}
