import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { localStorageStore } from "../../../../lib/store-utils/local-storage";
import { sessionStorageStore } from "../../../../lib/store-utils/session-storage";
import { get, writable } from "svelte/store";
import { waitUntil } from "@rilldata/web-common/lib/waitUtils.ts";

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

export const chatMounted = writable(false);

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

  startChat(prompt: string, submit = false): void {
    chatOpen.set(true);
    void waitUntil(() => get(chatMounted)).then(() =>
      eventBus.emit("start-chat", {
        prompt,
        submit,
      }),
    );
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
