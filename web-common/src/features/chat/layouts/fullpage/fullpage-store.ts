import { localStorageStore } from "../../../../lib/store-utils/local-storage";
import { sessionStorageStore } from "../../../../lib/store-utils/session-storage";

// =============================================================================
// FULLPAGE CONSTANTS
// =============================================================================

export const FULLPAGE_DEFAULTS = {
  CONVERSATION_SIDEBAR_OPEN: true, // Show conversation list by default
  CONVERSATION_SIDEBAR_WIDTH: 280,
  MIN_CONVERSATION_SIDEBAR_WIDTH: 240,
  MAX_CONVERSATION_SIDEBAR_WIDTH: 400,
} as const;

// =============================================================================
// FULLPAGE STORES
// =============================================================================

// Whether the conversation list sidebar is visible (for responsive behavior)
export const conversationSidebarOpen = sessionStorageStore<boolean>(
  "conversation-sidebar-open",
  FULLPAGE_DEFAULTS.CONVERSATION_SIDEBAR_OPEN,
);

// Width of the conversation sidebar (for potential future resizing)
export const conversationSidebarWidth = localStorageStore<number>(
  "conversation-sidebar-width",
  FULLPAGE_DEFAULTS.CONVERSATION_SIDEBAR_WIDTH,
);

// Search/filter state for conversations (future feature)
export const conversationSearchQuery = sessionStorageStore<string>(
  "conversation-search-query",
  "",
);

// =============================================================================
// FULLPAGE ACTIONS
// =============================================================================

export const fullpageActions = {
  toggleConversationSidebar(): void {
    conversationSidebarOpen.update((isOpen) => !isOpen);
  },

  openConversationSidebar(): void {
    conversationSidebarOpen.set(true);
  },

  closeConversationSidebar(): void {
    conversationSidebarOpen.set(false);
  },

  updateConversationSidebarWidth(width: number): void {
    const constrainedWidth = Math.max(
      FULLPAGE_DEFAULTS.MIN_CONVERSATION_SIDEBAR_WIDTH,
      Math.min(FULLPAGE_DEFAULTS.MAX_CONVERSATION_SIDEBAR_WIDTH, width),
    );
    conversationSidebarWidth.set(constrainedWidth);
  },

  setConversationSearchQuery(query: string): void {
    conversationSearchQuery.set(query);
  },

  clearConversationSearch(): void {
    conversationSearchQuery.set("");
  },
};
