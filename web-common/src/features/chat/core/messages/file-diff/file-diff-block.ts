import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "../../types";

// =============================================================================
// BACKEND TYPES (mirror runtime/ai tool definitions)
// =============================================================================

/** Arguments for the write_file tool call */
interface WriteFileCallData {
  path: string;
  contents: string;
}

/** Result from the write_file tool */
interface WriteFileResultData {
  diff?: string;
  is_new_file?: boolean;
  resources?: Array<{
    kind: string;
    name: string;
    reconcile_status: string;
    reconcile_error: string;
  }>;
  parse_error?: string;
  checkpoint_commit_hash?: string;
}

// =============================================================================
// BLOCK TYPE
// =============================================================================

/**
 * File diff block representation.
 * Contains a diff visualization extracted from a write_file tool call.
 */
export type FileDiffBlock = {
  type: "file-diff";
  id: string;
  message: V1Message;
  resultMessage: V1Message;
  filePath: string;
  diff: string;
  isNewFile: boolean;
  checkpointCommitHash: string | null;
};

/**
 * Creates a file diff block from a write_file tool call message.
 * Returns null if the data is invalid or the result indicates an error.
 */
export function createFileDiffBlock(
  message: V1Message,
  resultMessage: V1Message | undefined,
): FileDiffBlock | null {
  if (!resultMessage) return null;
  if (resultMessage.contentType === MessageContentType.ERROR) return null;

  try {
    const callData: WriteFileCallData = JSON.parse(message.contentData || "{}");
    const filePath = callData.path || "";
    if (!filePath) return null;

    const resultData: WriteFileResultData = JSON.parse(
      resultMessage.contentData || "{}",
    );

    return {
      type: "file-diff",
      id: `file-diff-${message.id}`,
      message,
      resultMessage,
      filePath,
      diff: resultData.diff || "",
      isNewFile: resultData.is_new_file || false,
      checkpointCommitHash: resultData.checkpoint_commit_hash || null,
    };
  } catch {
    return null;
  }
}
