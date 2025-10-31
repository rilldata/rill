import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { get } from "svelte/store";
import { localStorageStore } from "../../../../lib/store-utils/local-storage";
import { sessionStorageStore } from "../../../../lib/store-utils/session-storage";
import { getConversationManager } from "../../core/conversation-manager";

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

  startChat(instanceId: string, prompt: string): void {
    chatOpen.set(true);
    const conversationManager = getConversationManager(instanceId, {
      conversationState: "browserStorage",
    });
    const currentConversation = get(
      conversationManager.getCurrentConversation(),
    );
    currentConversation.draftMessage.set(prompt);
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
