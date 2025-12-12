import { localStorageStore } from "../../../../lib/store-utils/local-storage";

// =============================================================================
// SIDEBAR STATE
// =============================================================================

// Whether the conversation sidebar is collapsed (icon-only mode)
export const conversationSidebarCollapsed = localStorageStore<boolean>(
  "conversation-sidebar-collapsed",
  false, // default to expanded
);

// =============================================================================
// CONVERSATION ID PERSISTENCE
// =============================================================================

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
