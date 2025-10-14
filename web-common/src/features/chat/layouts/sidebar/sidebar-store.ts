import { type ConversationContextEntry } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import { getConversationManager } from "@rilldata/web-common/features/chat/core/conversation-manager.ts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { get } from "svelte/store";
import { localStorageStore } from "../../../../lib/store-utils/local-storage";
import { sessionStorageStore } from "../../../../lib/store-utils/session-storage";

// =============================================================================
// SIDEBAR CONSTANTS
// =============================================================================

export const SIDEBAR_DEFAULTS = {
  CHAT_OPEN: false,
  SIDEBAR_WIDTH: 500,
  MIN_SIDEBAR_WIDTH: 240,
  MAX_SIDEBAR_WIDTH: 600,
} as const;

// =============================================================================
// SIDEBAR STORES
// =============================================================================

export const chatOpen = sessionStorageStore<boolean>(
  "chat-open",
  SIDEBAR_DEFAULTS.CHAT_OPEN,
);

export const sidebarWidth = localStorageStore<number>(
  "sidebar-width",
  SIDEBAR_DEFAULTS.SIDEBAR_WIDTH,
);

// =============================================================================
// SIDEBAR ACTIONS
// =============================================================================

export const sidebarActions = {
  toggleChat(): void {
    chatOpen.update((isOpen) => !isOpen);
  },

  openChat(): void {
    chatOpen.set(true);
  },

  startChat(
    instanceId: string,
    prompt: string,
    contextEntries: ConversationContextEntry[],
  ): void {
    chatOpen.set(true);
    const conversationManager = getConversationManager(instanceId, {
      conversationState: "browserStorage",
    });
    conversationManager.enterNewConversationMode();
    get(conversationManager.getCurrentConversation()).draftMessage.set(prompt);
    get(conversationManager.getCurrentConversation()).context.override(
      contextEntries,
    );
    eventBus.emit("start-chat", null);
  },

  closeChat(): void {
    chatOpen.set(false);
  },

  updateSidebarWidth(width: number): void {
    const constrainedWidth = Math.max(
      SIDEBAR_DEFAULTS.MIN_SIDEBAR_WIDTH,
      Math.min(SIDEBAR_DEFAULTS.MAX_SIDEBAR_WIDTH, width),
    );
    sidebarWidth.set(constrainedWidth);
  },
};
