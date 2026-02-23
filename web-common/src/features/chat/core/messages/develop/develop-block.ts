import { type V1Message } from "@rilldata/web-common/runtime-client";
import {
  createFileDiffBlock,
  type FileDiffBlock,
} from "@rilldata/web-common/features/chat/core/messages/file-diff/file-diff-block.ts";
import {
  getLastMessage,
  getMessages,
  MessageSelectors,
} from "@rilldata/web-common/features/chat/core/messages/message-selectors.ts";
import {
  MessageType,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";
import type { RestoreChangesCallData } from "@rilldata/web-common/features/chat/core/messages/restore/restore-block.ts";

// =============================================================================
// BLOCK TYPE
// =============================================================================

export type DevelopBlock = {
  type: "develop";
  id: string;
  diffs: FileDiffBlock[];
  checkpointCommitHash: string;
  firstWriteCall?: V1Message;
  restored: boolean;
};

/**
 * Creates a file diff block from a write_file tool call message.
 * Returns null if the data is invalid or the result indicates an error.
 */
export function createDevelopBlock(
  writeMessages: V1Message[],
  id: string,
  resultMessagesByParentId: Map<string | undefined, V1Message>,
  messages: V1Message[],
): DevelopBlock | null {
  try {
    const diffs = writeMessages.map((message) =>
      createFileDiffBlock(message, resultMessagesByParentId.get(message.id)),
    );
    const nonNullDiffs = diffs.filter(Boolean) as FileDiffBlock[];
    if (nonNullDiffs.length === 0) return null;

    const checkpointCommitHash = diffs[0]?.checkpointCommitHash || "";

    const restoreCallMessages = getMessages(messages, [
      MessageSelectors.ByType(MessageType.CALL),
      MessageSelectors.ByToolName(ToolName.RESTORE_CHANGES),
    ]);
    const restored = restoreCallMessages.some((m) => {
      const restoredCall = JSON.parse(
        m.contentData || "{}",
      ) as RestoreChangesCallData;
      const restoredMessage = getLastMessage(messages, [
        MessageSelectors.ById(restoredCall.revert_till_write_call_id),
      ]);
      return restoredMessage
        ? restoredMessage.createdOn <= nonNullDiffs[0].message.createdOn
        : false;
    });

    return {
      type: "develop",
      id,
      diffs: nonNullDiffs,
      checkpointCommitHash,
      firstWriteCall: nonNullDiffs[0]?.message,
      restored,
    };
  } catch {
    return null;
  }
}
