/**
 * Conversation Selector - Abstracts how conversation selection state is managed and persisted
 *
 * This interface provides a clean abstraction for managing conversation selection
 * across different contexts (URL-based, browser storage, etc.) while keeping
 * the core chat logic agnostic to the specific storage mechanism.
 */

import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { derived, get, type Readable, type Writable } from "svelte/store";
import { sessionStorageStore } from "../../../lib/store-utils/session-storage";
import { NEW_CONVERSATION_ID } from "./utils";

// =============================================================================
// CORE INTERFACE
// =============================================================================

/**
 * Interface for managing conversation selection state
 *
 * Provides a clean, reactive interface for conversation selection with clear
 * intent-revealing methods and a single source of truth for current state.
 */
export interface ConversationSelector {
  /**
   * Single reactive source of truth for current conversation ID
   * NEW_CONVERSATION_ID indicates "new conversation" state
   */
  readonly currentConversationId: Readable<string>;

  /**
   * Navigate to a specific existing conversation
   */
  selectConversation(id: string): Promise<void>;

  /**
   * Clear current conversation selection (enter "new conversation" mode)
   */
  clearSelection(): Promise<void>;

  /**
   * Convenience reactive property for checking if in "new conversation" state
   */
  readonly isNewConversation: Readable<boolean>;
}

// =============================================================================
// URL CONVERSATION SELECTOR
// =============================================================================

/**
 * Manages conversation selection via URL parameters and browser navigation
 *
 * Used by FullPageChat where the URL is the source of truth for conversation selection.
 * Provides shareable URLs and browser back/forward support.
 *
 * Only handles REAL conversation IDs - optimistic state is managed by Chat class.
 */
export class URLConversationSelector implements ConversationSelector {
  // Single reactive source of truth
  readonly currentConversationId = derived(
    page,
    ($page) => $page.params.conversationId || NEW_CONVERSATION_ID,
  );

  // Derived convenience property
  readonly isNewConversation = derived(
    this.currentConversationId,
    ($id) => $id === NEW_CONVERSATION_ID,
  );

  async selectConversation(id: string): Promise<void> {
    const currentPage = get(page);
    const organization = currentPage.params.organization;
    const project = currentPage.params.project;
    await goto(`/${organization}/${project}/-/ai/${id}`, {
      replaceState: true,
    });
  }

  async clearSelection(): Promise<void> {
    const currentPage = get(page);
    const organization = currentPage.params.organization;
    const project = currentPage.params.project;
    await goto(`/${organization}/${project}/-/ai`, {
      replaceState: true,
    });
  }
}

// =============================================================================
// BROWSER STORAGE CONVERSATION SELECTOR
// =============================================================================

/**
 * Manages conversation selection via browser storage (sessionStorage/localStorage)
 *
 * Used by SidebarChat where conversation selection needs to persist across page
 * navigation but doesn't use URL routing. Selection is scoped per project.
 */
export class BrowserStorageConversationSelector
  implements ConversationSelector
{
  private store: Writable<string>;

  // Expose store as readonly reactive source
  readonly currentConversationId: Readable<string>;
  readonly isNewConversation: Readable<boolean>;

  constructor() {
    // Create project-specific storage store based on current page params
    const currentPage = get(page);
    const organization = currentPage.params.organization || "";
    const project = currentPage.params.project || "";

    this.store = sessionStorageStore<string>(
      `sidebar-conversation-id-${organization}-${project}`,
      NEW_CONVERSATION_ID,
    );

    // Expose as readonly reactive properties
    this.currentConversationId = { subscribe: this.store.subscribe };
    this.isNewConversation = derived(
      this.store,
      ($id) => $id === NEW_CONVERSATION_ID,
    );
  }

  async selectConversation(id: string): Promise<void> {
    this.store.set(id);
  }

  async clearSelection(): Promise<void> {
    this.store.set(NEW_CONVERSATION_ID);
  }

  /**
   * Optional: Handle external conversation updates
   * This allows the selector to sync when conversation selection changes externally
   */
  onConversationUpdate(conversationId: string): void {
    this.store.set(conversationId);
  }
}
