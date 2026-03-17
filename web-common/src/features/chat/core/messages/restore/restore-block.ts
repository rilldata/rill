// =============================================================================
// BACKEND TYPES (mirror runtime/ai tool definitions)
// =============================================================================

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "@rilldata/web-common/features/chat/core/types.ts";
import {
  getMessage,
  getNearestUserMessage,
  MessageSelectors,
} from "@rilldata/web-common/features/chat/core/messages/message-selectors.ts";

/** Arguments for the write_file tool call */
export interface RestoreChangesCallData {
  revert_till_write_call_id: string;
}

export type RestoreChangesBlock = {
  type: "restore-changes";
  id: string;
  restoredMessage: V1Message;
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

    const restoreCall = getMessage(allMessages, [
      MessageSelectors.ById(callData.revert_till_write_call_id),
    ]);
    if (!restoreCall) return null;

    const restoredMessage = getNearestUserMessage(allMessages, restoreCall);
    if (!restoredMessage) return null;

    return {
      type: "restore-changes",
      id: `restore-changes-${message.id}`,
      restoredMessage,
    };
  } catch {
    return null;
  }
}
