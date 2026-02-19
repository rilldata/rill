import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceListConversationsQueryOptions,
  type RpcStatus,
  type V1ListConversationsResponse,
} from "@rilldata/web-common/runtime-client";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, type Readable } from "svelte/store";
import { Conversation } from "./conversation";
import {
  BrowserStorageConversationSelector,
  URLConversationSelector,
  type ConversationSelector,
} from "./conversation-selector";
import { invalidateConversationsList, NEW_CONVERSATION_ID } from "./utils";

export type ConversationStateType = "url" | "browserStorage";

export interface ConversationManagerOptions {
  /**
   * How conversation state should be managed and persisted
   * - "url": Use URL parameters (for full-page chat with shareable URLs)
   * - "browserStorage": Use session storage (for sidebar chat)
   * - Ignored when `customSelector` is provided
   */
  conversationState: ConversationStateType;
  /**
   * The agent to use for conversations (e.g., "analyst_agent", "developer_agent")
   */
  agent?: string;
  /**
   * Optional custom conversation selector. When provided, `conversationState` is ignored.
   * Use this when the built-in selectors don't fit your URL structure (e.g., web-local /ai routes).
   */
  customSelector?: ConversationSelector;
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
  private newConversationUnsub: (() => void) | null = null;
  private conversations = new Map<string, Conversation>();
  private conversationSelector: ConversationSelector;
  private readonly agent?: string;

  constructor(
    public readonly instanceId: string,
    options: ConversationManagerOptions,
  ) {
    this.agent = options.agent;
    this.createNewConversation();

    if (options.customSelector) {
      this.conversationSelector = options.customSelector;
    } else {
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
          this.agent,
        );
        conversation.on("stream-start", () =>
          this.enforceMaxConcurrentStreams(),
        );
        conversation.on("conversation-forked", (newConversationId) =>
          this.handleConversationForked($conversationId, newConversationId),
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

  private createNewConversation() {
    this.newConversationUnsub?.();
    this.newConversation = new Conversation(
      this.instanceId,
      NEW_CONVERSATION_ID,
      this.agent,
    );
    const streamStartUnsub = this.newConversation.on("stream-start", () =>
      this.enforceMaxConcurrentStreams(),
    );
    const conversationStartedUnsub = this.newConversation.on(
      "conversation-created",
      (conversationId) => this.handleConversationCreated(conversationId),
    );
    this.newConversationUnsub = () => {
      streamStartUnsub();
      conversationStartedUnsub();
    };
  }

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
    this.conversationSelector.selectConversation(conversationId);
    void invalidateConversationsList(this.instanceId);
  }

  /**
   * Handle conversation forking - updates state to navigate to the forked conversation
   * Called when a non-owner sends a message and the conversation is forked
   */
  private handleConversationForked(
    originalConversationId: string,
    newConversationId: string,
  ): void {
    // Get the forked conversation (which is the updated instance)
    const forkedConversation = this.conversations.get(originalConversationId);
    if (forkedConversation) {
      // Move the conversation to the new ID in our map
      this.conversations.delete(originalConversationId);
      this.conversations.set(newConversationId, forkedConversation);
    }

    // Navigate to the new forked conversation
    this.conversationSelector.selectConversation(newConversationId);
  }

  /**
   * Rotates the new conversation: moves current "new" conversation to the conversations map
   * and creates a fresh "new" conversation instance
   */
  private rotateNewConversation(conversationId: string): void {
    // Store the new conversation instance in the conversations map
    this.conversations.set(conversationId, this.newConversation);

    // Create a fresh "new" conversation instance
    this.createNewConversation();
  }
}

// ===== CONVERSATION MANAGER SINGLETON MANAGEMENT =====

/**
 * Global registry of ConversationManager instances, one per instanceId+agent combination
 * Ensures consistent state across components within the same project and agent
 */
const conversationManagerInstances = new Map<string, ConversationManager>();

/**
 * Generate a unique key for conversation manager instances
 * @param instanceId - The project/instance identifier
 * @param agent - The agent type (e.g., "analyst_agent", "developer_agent")
 * @returns A unique key for the conversation manager
 */
function getConversationManagerKey(instanceId: string, agent?: string): string {
  return `${instanceId}:${agent || "default"}`;
}

/**
 * Get or create a ConversationManager instance for the given instanceId and agent
 *
 * @param instanceId - The project/instance identifier
 * @param options - Configuration options for the conversation manager instance
 * @returns The ConversationManager instance for this project and agent
 */
export function getConversationManager(
  instanceId: string,
  options: ConversationManagerOptions,
): ConversationManager {
  const key = getConversationManagerKey(instanceId, options.agent);
  if (!conversationManagerInstances.has(key)) {
    conversationManagerInstances.set(
      key,
      new ConversationManager(instanceId, options),
    );
  }
  return conversationManagerInstances.get(key)!;
}

/**
 * Clean up and remove a ConversationManager instance for the given instanceId
 * Called when leaving chat context or switching projects
 *
 * @param instanceId - The project/instance identifier to clean up
 */
export function cleanupConversationManager(instanceId: string): void {
  // Clean up all conversation managers for this instance (all agents)
  const keysToDelete: string[] = [];
  for (const key of conversationManagerInstances.keys()) {
    if (key.startsWith(`${instanceId}:`)) {
      const conversationManager = conversationManagerInstances.get(key);
      if (conversationManager) {
        conversationManager.cleanup();
      }
      keysToDelete.push(key);
    }
  }
  keysToDelete.forEach((key) => conversationManagerInstances.delete(key));
}
