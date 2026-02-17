/**
 * Local Conversation Selector - URL-based conversation selection for web-local
 *
 * This is a local version of URLConversationSelector that uses /ai/{id} URLs
 * instead of /{organization}/{project}/-/ai/{id}
 */

import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { derived } from "svelte/store";
import { NEW_CONVERSATION_ID } from "@rilldata/web-common/features/chat/core/utils";
import type { ConversationSelector } from "@rilldata/web-common/features/chat/core/conversation-selector";

export class LocalURLConversationSelector implements ConversationSelector {
  readonly currentConversationId = derived(
    page,
    ($page) => $page.params.conversationId || NEW_CONVERSATION_ID,
  );

  readonly isNewConversation = derived(
    this.currentConversationId,
    ($id) => $id === NEW_CONVERSATION_ID,
  );

  async selectConversation(id: string): Promise<void> {
    await goto(`/ai/${id}`, {
      replaceState: true,
    });
  }

  async clearSelection(): Promise<void> {
    await goto(`/ai`, {
      replaceState: true,
    });
  }
}
