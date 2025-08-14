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

function getConversationIdStorageKey(organization: string, project: string) {
  return `project-chat-conversation-id-${organization}-${project}`;
}

/**
 * Retrieves the last conversation ID for the given project from sessionStorage.
 * Handles both JSON-stringified and raw string formats for backwards compatibility.
 */
export function getLastConversationId(
  organization: string,
  project: string,
): string | null {
  const storageKey = getConversationIdStorageKey(organization, project);
  const storedValue = sessionStorage.getItem(storageKey);

  if (!storedValue) {
    return null;
  }

  try {
    // Try to parse as JSON first (new format)
    const parsed = JSON.parse(storedValue);
    return parsed === "null" ? null : parsed;
  } catch {
    // Fall back to raw string value (legacy format)
    return storedValue === "null" ? null : storedValue;
  }
}

/**
 * Stores the conversation ID for the given project in sessionStorage.
 * Uses JSON serialization for consistency with the getter function.
 */
export function setLastConversationId(
  organization: string,
  project: string,
  conversationId: string | null,
): void {
  const storageKey = getConversationIdStorageKey(organization, project);

  if (conversationId === null) {
    sessionStorage.removeItem(storageKey);
  } else {
    sessionStorage.setItem(storageKey, JSON.stringify(conversationId));
  }
}

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
};
