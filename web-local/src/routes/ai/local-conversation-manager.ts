/**
 * Local Conversation Manager - Uses web-common's ConversationManager with local URL selector
 *
 * Instead of duplicating ConversationManager, this thin wrapper injects
 * a LocalURLConversationSelector (which uses /ai/{id} URLs) via the
 * customSelector option.
 */

import {
  type ConversationManager,
  getConversationManager,
  cleanupConversationManager,
} from "@rilldata/web-common/features/chat/core/conversation-manager";
import { LocalURLConversationSelector } from "./local-conversation-selector";

export function getLocalConversationManager(
  instanceId: string,
  agent?: string,
): ConversationManager {
  return getConversationManager(instanceId, {
    conversationState: "url",
    agent,
    customSelector: new LocalURLConversationSelector(),
  });
}

export function cleanupLocalConversationManager(instanceId: string): void {
  cleanupConversationManager(instanceId);
}
