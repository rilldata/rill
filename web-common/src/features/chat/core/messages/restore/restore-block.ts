// =============================================================================
// BACKEND TYPES (mirror runtime/ai tool definitions)
// =============================================================================

import type { V1Message } from "@rilldata/web-common/runtime-client";
import {
  MessageContentType,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";

/** Arguments for the write_file tool call */
interface RestoreChangesCallData {
  revert_till_write_call_id: string;
}

export type RestoreChangesBlock = {
  type: "restore-changes";
  id: string;
  revertedConversations: V1Message[];
};

export function createRestoreChangesBlock(
  message: V1Message,
  _: V1Message | undefined,
  allMessages: V1Message[],
): RestoreChangesBlock | null {
  if (!message) return null;
  if (message.contentType === MessageContentType.ERROR) return null;

  try {
    const callData: RestoreChangesCallData = JSON.parse(
      message.contentData || "{}",
    );
    if (!callData.revert_till_write_call_id) return null;
    const revertedConversations: V1Message[] = [];
    let found = false;
    for (let i = allMessages.length - 1; i >= 0; i--) {
      const message = allMessages[i];
      if (message.id === callData.revert_till_write_call_id) found = true;

      if (message.role === "user") {
        revertedConversations.push(message);
        if (found) break;
      }
    }

    return {
      type: "restore-changes",
      id: `restore-changes-${message.id}`,
      revertedConversations,
    };
  } catch {
    return null;
  }
}
