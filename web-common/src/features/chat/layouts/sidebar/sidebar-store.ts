import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import { localStorageStore } from "../../../../lib/store-utils/local-storage";
import { sessionStorageStore } from "../../../../lib/store-utils/session-storage";
import { get, writable, type Writable } from "svelte/store";
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

// Per-surface open state. Keeping the developer and dashboard panels on
// independent keys means publishing from a Rill Developer tab does not flip
// the chat-open flag in the freshly opened production tab — Chromium clones
// sessionStorage when window.open inherits the opener context.
export const developerChatOpen = sessionStorageStore<boolean>(
  "chat-open-developer",
  SIDEBAR_DEFAULTS.CHAT_OPEN,
);

export const dashboardChatOpen = sessionStorageStore<boolean>(
  "chat-open-dashboard",
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

export type ChatActions = {
  toggleChat(): void;
  openChat(): void;
  closeChat(): void;
  startChat(prompt: string): void;
  updateSidebarWidth(width: number): void;
};

function createChatActions(open: Writable<boolean>): ChatActions {
  return {
    toggleChat() {
      open.update((isOpen) => !isOpen);
    },
    openChat() {
      open.set(true);
    },
    closeChat() {
      open.set(false);
    },
    startChat(prompt: string) {
      open.set(true);
      void waitUntil(() => get(chatMounted)).then(() =>
        eventBus.emit("start-chat", prompt),
      );
    },
    updateSidebarWidth(width: number) {
      const constrainedWidth = Math.max(
        SIDEBAR_DEFAULTS.MIN_SIDEBAR_WIDTH,
        Math.min(SIDEBAR_DEFAULTS.MAX_SIDEBAR_WIDTH, width),
      );
      sidebarWidth.set(constrainedWidth);
    },
  };
}

export const developerChatActions = createChatActions(developerChatOpen);
export const dashboardChatActions = createChatActions(dashboardChatOpen);
