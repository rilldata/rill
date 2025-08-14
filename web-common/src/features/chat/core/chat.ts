import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetConversationQueryOptions,
  getRuntimeServiceListConversationsQueryOptions,
  type RpcStatus,
  type V1GetConversationResponse,
  type V1ListConversationsResponse,
} from "@rilldata/web-common/runtime-client";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";
import { NEW_CONVERSATION_ID } from "./chat-utils";
import { Conversation } from "./conversation";
import {
  BrowserStorageConversationSelector,
  URLConversationSelector,
  type ConversationSelector,
} from "./conversation-selector";

export type ConversationStateType = "url" | "browserStorage";

export interface ChatOptions {
  /**
   * How conversation state should be managed and persisted
   * - "url": Use URL parameters (for full-page chat with shareable URLs)
   * - "browserStorage": Use session storage (for sidebar chat)
   */
  conversationState: ConversationStateType;
}

/**
 * Manages chat state and conversation lifecycle.
 *
 * Provides reactive stores for conversation lists and current conversation selection.
 * Handles conversation creation, navigation, and cleanup across different UI contexts
 * (full-page chat with URL state vs sidebar chat with browser storage).
 *
 * Usage:
 * - Access conversations: `chat.listConversationsQuery()` and `chat.getCurrentConversation()`
 * - Navigate conversations: `chat.selectConversation(id)` or `chat.enterNewConversationMode()`
 * - Send messages: `conversation.sendMessage()` on any conversation instance (new or existing)
 */
export class Chat {
  private newConversation: Conversation;
  private conversations = new Map<string, Conversation>();
  private conversationSelector: ConversationSelector;

  constructor(
    private instanceId: string,
    options: ChatOptions,
  ) {
    this.newConversation = new Conversation(
      this.instanceId,
      NEW_CONVERSATION_ID,
      {
        onConversationCreated: (conversationId: string) => {
          this.selectConversation(conversationId);
        },
      },
    );

    switch (options.conversationState) {
      case "url":
        this.conversationSelector = new URLConversationSelector();
        break;
      case "browserStorage":
        this.conversationSelector = new BrowserStorageConversationSelector();
        break;
      default:
        throw new Error(
          `Unknown conversation storage type: ${options.conversationState}`,
        );
    }
  }

  // QUERIES

  listConversationsQuery(): CreateQueryResult<
    V1ListConversationsResponse,
    RpcStatus
  > {
    return createQuery(
      getRuntimeServiceListConversationsQueryOptions(this.instanceId),
      queryClient,
    );
  }

  getCurrentConversation(): Readable<Conversation> {
    return derived(
      [this.conversationSelector.currentConversationId],
      ([$conversationId]) => {
        // If the conversation ID is "new", return the new conversation instance
        if ($conversationId === NEW_CONVERSATION_ID) {
          return this.newConversation;
        }

        // If we already have a conversation instance for this conversation ID, return it
        const existing = this.conversations.get($conversationId);
        if (existing) return existing;

        // Otherwise, create a conversation instance and store it
        const conversation = new Conversation(this.instanceId, $conversationId);
        this.conversations.set($conversationId, conversation);
        return conversation;
      },
    );
  }

  // This method is an optimization for getting a `GetConversation` QueryObserver for the current conversation
  // However, this admittedly blurs the boundary between the `Chat` class and the `Conversation` class (it's "feature creep")
  getCurrentConversationQuery(): CreateQueryResult<
    V1GetConversationResponse,
    RpcStatus
  > {
    const queryOptions = derived(
      [this.conversationSelector.currentConversationId],
      ([$conversationId]) =>
        getRuntimeServiceGetConversationQueryOptions(
          this.instanceId,
          $conversationId,
        ),
    );

    return createQuery(queryOptions, queryClient);
  }

  // ACTIONS

  enterNewConversationMode(): void {
    this.conversationSelector.clearSelection();
  }

  selectConversation(conversationId: string): void {
    this.conversationSelector.selectConversation(conversationId);
  }
}
