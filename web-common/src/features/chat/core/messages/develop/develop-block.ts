import {
  type V1Message,
  type V1ResourceName,
} from "@rilldata/web-common/runtime-client";
import {
  createFileDiffBlock,
  type FileDiffBlock,
} from "@rilldata/web-common/features/chat/core/messages/file-diff/file-diff-block.ts";

// =============================================================================
// BLOCK TYPE
// =============================================================================

export type DevelopBlock = {
  type: "develop";
  id: string;
  diffs: FileDiffBlock[];
  checkpointCommitHash: string;
};

/**
 * Creates a file diff block from a write_file tool call message.
 * Returns null if the data is invalid or the result indicates an error.
 */
export function createDevelopBlock(
  writeMessages: V1Message[],
  id: string,
  resultMessagesByParentId: Map<string | undefined, V1Message>,
): DevelopBlock | null {
  try {
    const diffs = writeMessages.map((message) =>
      createFileDiffBlock(message, resultMessagesByParentId.get(message.id)),
    );

    return {
      type: "develop",
      id,
      diffs: diffs.filter(Boolean) as FileDiffBlock[],
      checkpointCommitHash: diffs[0]?.checkpointCommitHash || "",
    };
  } catch {
    return null;
  }
}
