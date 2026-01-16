// =============================================================================
// BACKEND TYPES (mirror runtime/ai tool definitions)
// =============================================================================

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "@rilldata/web-common/features/chat/core/types.ts";
import {
  createFileDiffBlock,
  type FileDiffBlock,
} from "@rilldata/web-common/features/chat/core/messages/file-diff/file-diff-block.ts";

/** Result from the developer_agent tool */
interface DeveloperAgentResultData {
  response: string;
  commit_hash: string;
}

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
  resultMessage: V1Message | undefined,
  id: string,
  resultMessagesByParentId: Map<string | undefined, V1Message>,
): DevelopBlock | null {
  if (!resultMessage) return null;
  if (resultMessage.contentType === MessageContentType.ERROR) return null;

  try {
    const resultData: DeveloperAgentResultData = JSON.parse(
      resultMessage.contentData || "{}",
    );

    const diffs = writeMessages.map((message) =>
      createFileDiffBlock(message, resultMessagesByParentId.get(message.id)),
    );

    return {
      type: "develop",
      id,
      diffs: diffs.filter(Boolean) as FileDiffBlock[],
      checkpointCommitHash: resultData.commit_hash || "",
    };
  } catch {
    return null;
  }
}
