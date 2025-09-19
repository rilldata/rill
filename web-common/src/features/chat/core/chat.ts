import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetConversationQueryKey,
  getRuntimeServiceListConversationsQueryKey,
  getRuntimeServiceListConversationsQueryOptions,
  type RpcStatus,
  type V1Conversation,
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
          this.handleConversationCreated(conversationId);
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

  // ACTIONS

  enterNewConversationMode(): void {
    this.conversationSelector.clearSelection();
  }

  selectConversation(conversationId: string): void {
    this.conversationSelector.selectConversation(conversationId);
  }

  /**
   * Handle conversation creation - rotates conversation instances, updates list cache, and navigates to it
   */
  private handleConversationCreated(conversationId: string): void {
    this.rotateNewConversation(conversationId);
    this.updateConversationListCache(conversationId);
    this.conversationSelector.selectConversation(conversationId);
  }

  /**
   * Rotates the new conversation: moves current "new" conversation to the conversations map
   * and creates a fresh "new" conversation instance
   */
  private rotateNewConversation(conversationId: string): void {
    // Store the new conversation instance in the map of existing conversations
    this.conversations.set(conversationId, this.newConversation);

    // Create a fresh "new" conversation instance
    this.newConversation = new Conversation(
      this.instanceId,
      NEW_CONVERSATION_ID,
      {
        onConversationCreated: (conversationId: string) => {
          this.handleConversationCreated(conversationId);
        },
      },
    );
  }

  /**
   * Update the conversation list cache by adding the new conversation
   */
  private updateConversationListCache(conversationId: string): void {
    const listConversationsKey = getRuntimeServiceListConversationsQueryKey(
      this.instanceId,
    );

    queryClient.setQueryData<V1ListConversationsResponse>(
      listConversationsKey,
      (old) => {
        const conversations = old?.conversations ?? [];

        // Check if conversation already exists in the list
        const existingIndex = conversations.findIndex(
          (c) => c.id === conversationId,
        );
        if (existingIndex >= 0) {
          // Conversation already exists, no need to add it again
          return old;
        }

        // Fetch conversation data from the GetConversation query cache
        const conversationCacheKey = getRuntimeServiceGetConversationQueryKey(
          this.instanceId,
          conversationId,
        );
        const cachedConversationData =
          queryClient.getQueryData<V1GetConversationResponse>(
            conversationCacheKey,
          );
        const conversationData = cachedConversationData?.conversation;

        // Create conversation object for the list
        const newConversation: V1Conversation = {
          id: conversationId,
          title: this.generateConversationTitle(conversationData),
          createdOn: conversationData?.createdOn || new Date().toISOString(),
          updatedOn: conversationData?.updatedOn || new Date().toISOString(),
          messages: [], // Don't include messages in the list view
        };

        // Add the new conversation to the front of the list
        conversations.unshift(newConversation);
        return { ...old, conversations };
      },
    );
  }

  /**
   * Generate a conversation title from the conversation data
   */
  private generateConversationTitle(conversationData?: V1Conversation): string {
    // If we have conversation data with messages, generate title from first user message
    if (conversationData?.messages) {
      for (const message of conversationData.messages) {
        if (message.role === "user" && message.content?.[0]?.text) {
          let title = message.content[0].text.trim();

          // Truncate to 50 characters and add ellipsis if needed
          if (title.length > 50) {
            title = title.substring(0, 50) + "...";
          }

          // Replace newlines with spaces and collapse multiple spaces
          title = title.replace(/[\r\n]/g, " ").replace(/\s+/g, " ");

          return title;
        }
      }
    }

    // Fallback title
    return "New conversation";
  }
}

// ===== CHAT SINGLETON MANAGEMENT =====

// Chat instance singleton map
const chatInstances = new Map<string, Chat>();

/**
 * Get or create a Chat instance for the given instanceId
 */
export function getChatInstance(
  instanceId: string,
  options: ChatOptions,
): Chat {
  if (!chatInstances.has(instanceId)) {
    chatInstances.set(instanceId, new Chat(instanceId, options));
  }
  return chatInstances.get(instanceId)!;
}
