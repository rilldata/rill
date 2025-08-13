import { page } from "$app/stores";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetConversationQueryOptions,
  getRuntimeServiceListConversationsQueryKey,
  getRuntimeServiceListConversationsQueryOptions,
  runtimeServiceComplete,
  type RpcStatus,
  type V1GetConversationResponse,
  type V1ListConversationsResponse,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, writable, type Readable } from "svelte/store";
import { detectAppContext, getOptimisticMessageId } from "./chat-utils";
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
 * - Start new conversation: `chat.createConversation(message)`
 * - Navigate conversations: `chat.selectConversation(id)` or `chat.enterNewConversationMode()`
 * - Send to existing conversation: `conversation.sendMessage()` on the conversation instance
 */
export class Chat {
  public pendingMessage = writable<V1Message | null>(null);
  private conversationSelector: ConversationSelector;
  private conversations = new Map<string, Conversation>();

  constructor(
    private instanceId: string,
    options: ChatOptions,
  ) {
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

  getCurrentConversation(): Readable<Conversation | null> {
    return derived(
      [this.conversationSelector.currentConversationId],
      ([$conversationId]) => {
        if (!$conversationId) return null;

        // Check if we already have a conversation instance for this conversation ID
        const existing = this.conversations.get($conversationId);
        if (existing) return existing;

        // If not, create a new conversation instance and store it
        const conversation = Conversation.fromExistingId(
          this.instanceId,
          $conversationId,
        );
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
   * Creates a new conversation by sending the first message.
   * Handles optimistic updates for immediate UI feedback.
   */
  async createConversation(initialMessage: string): Promise<void> {
    if (!initialMessage.trim()) {
      throw new Error("Cannot start a conversation with an empty message");
    }

    this.pendingMessage.set({
      id: getOptimisticMessageId(),
      role: "user" as const,
      content: [{ text: initialMessage }],
      createdOn: new Date().toISOString(),
      updatedOn: new Date().toISOString(),
    });

    try {
      // Hit the `Complete` API with no `conversationId` to create a new conversation
      const response = await runtimeServiceComplete(this.instanceId, {
        conversationId: undefined, // New conversation
        messages: [
          {
            role: "user" as const,
            content: [{ text: initialMessage }],
          },
        ],
        appContext: detectAppContext(get(page)) ?? undefined,
      });

      if (!response.conversationId) {
        throw new Error("Did not receive a conversation ID from the server.");
      }

      const realConversationId = response.conversationId;

      // Transition optimistic state to real conversation

      // Fetch complete conversation data (including the conversation title)
      const conversationResponse =
        await queryClient.fetchQuery<V1GetConversationResponse>(
          getRuntimeServiceGetConversationQueryOptions(
            this.instanceId,
            realConversationId,
          ),
        );

      const finalConversation = conversationResponse.conversation;
      if (!finalConversation) {
        throw new Error("Could not fetch the conversation.");
      }

      // Update conversation list cache
      const listConversationsKey = getRuntimeServiceListConversationsQueryKey(
        this.instanceId,
      );
      queryClient.setQueryData<V1ListConversationsResponse>(
        listConversationsKey,
        (old) => {
          const conversations = old?.conversations ?? [];
          conversations.unshift(finalConversation);
          return { ...old, conversations };
        },
      );

      // Switch to the real conversation
      await this.conversationSelector.selectConversation(realConversationId);
    } catch (error) {
      console.error("Failed to start new conversation:", error);
      throw error;
    } finally {
      // Clear optimistic state
      this.pendingMessage.set(null);
    }
  }
}
