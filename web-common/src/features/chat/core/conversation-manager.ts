import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetConversationQueryKey,
  getRuntimeServiceListConversationsQueryKey,
  getRuntimeServiceListConversationsQueryOptions,
  type RpcStatus,
  type V1Conversation,
  type V1GetConversationResponse,
  type V1ListConversationsResponse,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, type Readable } from "svelte/store";
import { Conversation } from "./conversation";
import {
  BrowserStorageConversationSelector,
  URLConversationSelector,
  type ConversationSelector,
} from "./conversation-selector";
import { extractMessageText, NEW_CONVERSATION_ID } from "./utils";

export type ConversationStateType = "url" | "browserStorage";

export interface ConversationManagerOptions {
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
 * - Access conversations: `conversationManager.listConversationsQuery()` and `conversationManager.getCurrentConversation()`
 * - Navigate conversations: `conversationManager.selectConversation(id)` or `conversationManager.enterNewConversationMode()`
 * - Send messages: `conversation.sendMessage()` on any conversation instance (new or existing)
 */
export class ConversationManager {
  // Maximum number of conversations that can have active streaming at once
  private static readonly MAX_CONCURRENT_STREAMS = 3;

  private newConversation: Conversation;
  private conversations = new Map<string, Conversation>();
  private conversationSelector: ConversationSelector;

  constructor(
    private instanceId: string,
    options: ConversationManagerOptions,
  ) {
    this.newConversation = new Conversation(
      this.instanceId,
      NEW_CONVERSATION_ID,
      {
        onStreamStart: () => this.enforceMaxConcurrentStreams(),
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

  // ===== PUBLIC API =====

  /**
   * Get a reactive query for the list of conversations
   */
  public listConversationsQuery(): CreateQueryResult<
    V1ListConversationsResponse,
    RpcStatus
  > {
    return createQuery(
      getRuntimeServiceListConversationsQueryOptions(this.instanceId, {
        // Filter to only show Rill client conversations, excluding MCP conversations
        userAgentPattern: "rill%",
      }),
      queryClient,
    );
  }

  /**
   * Get a reactive store for the currently selected conversation
   */
  public getCurrentConversation(): Readable<Conversation> {
    return derived(
      [this.conversationSelector.currentConversationId],
      ([$conversationId]) => {
        // If the conversation ID is "new", return the new conversation instance
        if ($conversationId === NEW_CONVERSATION_ID) {
          return this.newConversation;
        }

        // If we already have a conversation instance for this conversation ID, return it
        const existing = this.conversations.get($conversationId);
        if (existing) {
          return existing;
        }

        // Otherwise, create a conversation instance and store it
        const conversation = new Conversation(
          this.instanceId,
          $conversationId,
          {
            onStreamStart: () => this.enforceMaxConcurrentStreams(),
          },
        );
        this.conversations.set($conversationId, conversation);
        return conversation;
      },
    );
  }

  /**
   * Navigate to a specific conversation
   */
  public selectConversation(conversationId: string): void {
    this.conversationSelector.selectConversation(conversationId);
  }

  /**
   * Enter new conversation mode (clear current selection)
   */
  public enterNewConversationMode(): void {
    this.conversationSelector.clearSelection();
  }

  /**
   * Clean up all resources for this Chat instance
   */
  public cleanup(): void {
    // Clean up all conversation instances
    this.conversations.forEach((conversation) => {
      conversation.cleanup();
    });
    this.conversations.clear();

    // Clean up the new conversation instance
    this.newConversation.cleanup();
  }

  // ===== PRIVATE IMPLEMENTATION =====

  // ----- Stream Management -----

  /**
   * Get conversations that are currently streaming
   */
  private getActiveStreamingConversations(): Conversation[] {
    return Array.from(this.conversations.values()).filter((conv) =>
      get(conv.isStreaming),
    );
  }

  /**
   * Stop streaming in oldest conversations if we exceed concurrent stream limit
   */
  private enforceMaxConcurrentStreams(): void {
    try {
      const streamingConversations = this.getActiveStreamingConversations();

      if (
        streamingConversations.length >=
        ConversationManager.MAX_CONCURRENT_STREAMS
      ) {
        // Stop the oldest streaming conversations (simple FIFO approach)
        const conversationsToStop = streamingConversations.slice(
          0,
          streamingConversations.length -
            ConversationManager.MAX_CONCURRENT_STREAMS +
            1,
        );

        conversationsToStop.forEach((conv) => {
          conv.cancelStream();
        });
      }
    } catch (error) {
      console.warn("Error enforcing max concurrent streams:", error);
    }
  }

  // ----- Conversation Lifecycle -----

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
    // Store the new conversation instance in the conversations map
    this.conversations.set(conversationId, this.newConversation);

    // Create a fresh "new" conversation instance
    this.newConversation = new Conversation(
      this.instanceId,
      NEW_CONVERSATION_ID,
      {
        onStreamStart: () => this.enforceMaxConcurrentStreams(),
        onConversationCreated: (conversationId: string) => {
          this.handleConversationCreated(conversationId);
        },
      },
    );
  }

  // ----- Cache Management -----

  /**
   * Update the conversation list cache by adding the new conversation
   */
  private updateConversationListCache(conversationId: string): void {
    const listConversationsKey = getRuntimeServiceListConversationsQueryKey(
      this.instanceId,
      {
        userAgentPattern: "rill%",
      },
    );

    // Check if we have existing cached data
    const existingData =
      queryClient.getQueryData<V1ListConversationsResponse>(
        listConversationsKey,
      );

    // If no cached data exists, invalidate to fetch fresh data instead of creating an empty list
    if (!existingData) {
      queryClient.invalidateQueries({ queryKey: listConversationsKey });
      return;
    }

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
        const cachedGetConversationResponse =
          queryClient.getQueryData<V1GetConversationResponse>(
            conversationCacheKey,
          ) as V1GetConversationResponse | undefined;
        const conversation = cachedGetConversationResponse?.conversation;

        // Create conversation object for the list
        const newConversation: V1Conversation = {
          id: conversationId,
          title: this.generateConversationTitle(
            cachedGetConversationResponse?.messages,
          ),
          createdOn: conversation?.createdOn || new Date().toISOString(),
          updatedOn: conversation?.updatedOn || new Date().toISOString(),
        };

        // Add the new conversation to the front of the list
        conversations.unshift(newConversation);
        return { ...old, conversations };
      },
    );
  }

  /**
   * Generate a conversation title from messages
   *
   * Note: This replicates the server-side title generation logic client-side
   * to avoid making an additional network request for something we can compute
   * trivially from the conversation data we already have in cache.
   */
  private generateConversationTitle(messages?: V1Message[]): string {
    // If we have messages, generate title from first user message
    if (messages) {
      for (const message of messages) {
        if (message.role === "user") {
          let title = extractMessageText(message);

          if (!title) continue;

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

// ===== CONVERSATION MANAGER SINGLETON MANAGEMENT =====

/**
 * Global registry of ConversationManager instances, one per instanceId (project)
 * Ensures consistent state across components within the same project
 */
const conversationManagerInstances = new Map<string, ConversationManager>();

/**
 * Get or create a ConversationManager instance for the given instanceId
 *
 * @param instanceId - The project/instance identifier
 * @param options - Configuration options for the conversation manager instance
 * @returns The ConversationManager instance for this project
 */
export function getConversationManager(
  instanceId: string,
  options: ConversationManagerOptions,
): ConversationManager {
  if (!conversationManagerInstances.has(instanceId)) {
    conversationManagerInstances.set(
      instanceId,
      new ConversationManager(instanceId, options),
    );
  }
  return conversationManagerInstances.get(instanceId)!;
}

/**
 * Clean up and remove a ConversationManager instance for the given instanceId
 * Called when leaving chat context or switching projects
 *
 * @param instanceId - The project/instance identifier to clean up
 */
export function cleanupConversationManager(instanceId: string): void {
  const conversationManager = conversationManagerInstances.get(instanceId);
  if (conversationManager) {
    conversationManager.cleanup();
    conversationManagerInstances.delete(instanceId);
  }
}
