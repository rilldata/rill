import { ConversationContext } from "@rilldata/web-common/features/chat/core/context/context.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetConversationQueryKey,
  getRuntimeServiceGetConversationQueryOptions,
  type RpcStatus,
  type V1CompleteStreamingResponse,
  type V1GetConversationResponse,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { SSEFetchClient } from "@rilldata/web-common/runtime-client/sse-fetch-client";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, writable, type Readable } from "svelte/store";
import type { HTTPError } from "../../../runtime-client/fetchWrapper";
import {
  formatChatError,
  getOptimisticMessageId,
  NEW_CONVERSATION_ID,
} from "./chat-utils";

/**
 * Individual conversation state management.
 *
 * Handles streaming message sending, optimistic updates, and conversation-specific queries
 * for a single conversation using the streaming completion endpoint.
 */
export class Conversation {
  // Public reactive state
  public readonly draftMessage = writable<string>("");
  public readonly isStreaming = writable(false);
  public readonly streamError = writable<string | null>(null);
  public readonly context = new ConversationContext();

  // Private state
  private sseClient: SSEFetchClient<V1CompleteStreamingResponse> | null = null;
  private hasReceivedFirstUserMessage = false;

  constructor(
    private readonly instanceId: string,
    public conversationId: string,
    private readonly options?: {
      onStreamStart?: () => void;
      onConversationCreated?: (conversationId: string) => void;
    },
  ) {}

  // ===== PUBLIC API =====

  /**
   * Get a reactive query for this conversation's data
   */
  public getConversationQuery(): CreateQueryResult<
    V1GetConversationResponse,
    RpcStatus
  > {
    return createQuery(
      getRuntimeServiceGetConversationQueryOptions(
        this.instanceId,
        this.conversationId,
        undefined,
        {
          query: {
            enabled: this.conversationId !== NEW_CONVERSATION_ID,
          },
        },
      ),
      queryClient,
    );
  }

  /**
   * Get a reactive store for conversation query errors
   */
  public getConversationQueryError(): Readable<string | null> {
    return derived(
      this.getConversationQuery(),
      ($getConversationQuery) =>
        ($getConversationQuery.error as HTTPError)?.response?.data?.message ??
        null,
    );
  }

  /**
   * Send a message and handle streaming response
   *
   * @param options - Callback functions for different stages of message sending
   */
  public async sendMessage(options?: {
    onStreamStart?: () => void;
    onMessage?: (message: V1Message) => void;
    onStreamComplete?: (conversationId: string) => void;
    onError?: (error: string) => void;
  }): Promise<void> {
    // Prevent concurrent message sending
    if (get(this.isStreaming)) {
      this.streamError.set("Please wait for the current response to complete");
      return;
    }

    let prompt = get(this.draftMessage).trim();
    if (!prompt) throw new Error("Cannot send empty message");
    prompt += this.context.toString();

    // Optimistic updates
    this.draftMessage.set("");
    this.streamError.set(null);
    this.isStreaming.set(true);
    this.hasReceivedFirstUserMessage = false; // Reset for new message

    const userMessage = this.addOptimisticUserMessage(prompt);

    try {
      // Start streaming - this establishes the connection
      const streamPromise = this.startStreaming(prompt);

      // Stream has been initiated, notify caller
      options?.onStreamStart?.();

      // Wait for streaming to complete
      await streamPromise;

      // Stream has completed successfully
      options?.onStreamComplete?.(this.conversationId);
    } catch (error) {
      // Business logic errors: flow control, validation, business failures
      // These require full rollback of optimistic updates
      this.handleStreamingError(error, userMessage);
      options?.onError?.(formatChatError(error));
    } finally {
      this.isStreaming.set(false);
    }
  }

  /**
   * Cancel the current streaming session (user-initiated)
   */
  public cancelStream(): void {
    if (this.sseClient) {
      this.sseClient.stop();
    }
    this.isStreaming.set(false);
    this.streamError.set(null);
  }

  /**
   * Clean up all resources when conversation is no longer needed (lifecycle cleanup)
   */
  public cleanup(): void {
    // Cancel any active streaming first
    this.cancelStream();

    this.context.clear();

    // Full resource cleanup
    if (this.sseClient) {
      this.sseClient.cleanup();
      this.sseClient = null;
    }
  }

  // ===== PRIVATE IMPLEMENTATION =====

  // ----- Transport Layer: SSE Connection Management -----

  /**
   * Start streaming completion responses for a given prompt
   * Returns a Promise that resolves when streaming completes
   */
  private async startStreaming(prompt: string): Promise<void> {
    // Initialize SSE client if not already done
    if (!this.sseClient) {
      this.sseClient = new SSEFetchClient<V1CompleteStreamingResponse>();

      // Set up transport-level event handlers
      this.sseClient.on("data", (response) => {
        // Delegate to business logic handler
        this.processStreamingResponse(response);
      });

      this.sseClient.on("error", (error) => {
        // Transport errors: connection, parsing, network issues
        // No rollback needed - these are infrastructure failures
        this.streamError.set(this.getDescriptiveError(error));
      });

      this.sseClient.on("close", () => {
        // Transport completed - business logic handles completion in sendMessage
      });
    }

    // Clean up any existing connection
    this.sseClient.stop();

    // Build URL with stream parameter (like other streaming endpoints)
    const baseUrl = `${get(runtime).host}/v1/instances/${this.instanceId}/ai/complete/stream?stream=messages`;

    // Prepare request body for POST request
    const requestBody = {
      instanceId: this.instanceId,
      conversationId:
        this.conversationId === NEW_CONVERSATION_ID
          ? undefined
          : this.conversationId,
      prompt,
    };

    // Notify that streaming is about to start (for concurrent stream management)
    this.options?.onStreamStart?.();

    // Start streaming - this will establish the connection and then stream until completion
    const streamPromise = this.sseClient.start(baseUrl, {
      method: "POST",
      body: requestBody,
    });

    // Stream has been initiated - we can notify that streaming started
    // (the connection is established at this point, even though data may still be streaming)

    // Wait for the stream to complete
    await streamPromise;
  }

  // ----- Business Logic Layer: Message Processing -----

  /**
   * Process streaming response data and update conversation state
   * This handles the business logic for each streaming message
   */
  private processStreamingResponse(
    response: V1CompleteStreamingResponse,
  ): void {
    // Handle conversation ID transition for new conversations
    if (
      response.conversationId &&
      response.conversationId !== this.conversationId &&
      this.conversationId === NEW_CONVERSATION_ID
    ) {
      this.transitionToRealConversation(response.conversationId);
    }

    if (response.message) {
      // Skip the first user message from the stream since we've already added it optimistically
      if (
        response.message.role === "user" &&
        !this.hasReceivedFirstUserMessage
      ) {
        this.hasReceivedFirstUserMessage = true;
        return;
      }

      // Check if this is a tool result message
      const toolResult = response.message.content?.[0]?.toolResult;
      if (toolResult?.id) {
        this.handleToolResult(toolResult);
      } else {
        this.addMessageToCache(response.message);
      }
    }
  }

  // ----- Conversation Lifecycle -----

  /**
   * Transition from NEW_CONVERSATION_ID to real conversation ID
   * Transfers all cached data to the new conversation cache
   */
  private transitionToRealConversation(realConversationId: string): void {
    const oldCacheKey = getRuntimeServiceGetConversationQueryKey(
      this.instanceId,
      this.conversationId, // This is still "new"
    );

    const newCacheKey = getRuntimeServiceGetConversationQueryKey(
      this.instanceId,
      realConversationId,
    );

    // Get existing data from "new" conversation cache
    const existingData =
      queryClient.getQueryData<V1GetConversationResponse>(oldCacheKey);

    if (existingData?.conversation) {
      // Transfer the conversation data to the real conversation ID cache
      queryClient.setQueryData<V1GetConversationResponse>(newCacheKey, {
        conversation: {
          ...existingData.conversation,
          id: realConversationId,
        },
      });
    }

    // Clean up the old "new" conversation cache
    queryClient.removeQueries({ queryKey: oldCacheKey });

    // Update the conversation ID
    this.conversationId = realConversationId;

    // Notify that conversation was created
    this.options?.onConversationCreated?.(realConversationId);
  }

  // ----- Cache Management -----

  /**
   * Add optimistic user message to cache
   */
  private addOptimisticUserMessage(prompt: string): V1Message {
    const userMessage: V1Message = {
      id: getOptimisticMessageId(),
      role: "user",
      content: [{ text: prompt }],
      createdOn: new Date().toISOString(),
      updatedOn: new Date().toISOString(),
    };

    this.addMessageToCache(userMessage);
    return userMessage;
  }

  /**
   * Add or merge message to TanStack Query cache
   */
  private addMessageToCache(message: V1Message): void {
    const cacheKey = getRuntimeServiceGetConversationQueryKey(
      this.instanceId,
      this.conversationId,
    );
    queryClient.setQueryData<V1GetConversationResponse>(cacheKey, (old) => {
      if (!old?.conversation) {
        // Create initial conversation structure if it doesn't exist
        return {
          conversation: {
            id: this.conversationId,
            messages: [message],
            createdOn: message.createdOn,
            updatedOn: new Date().toISOString(),
          },
        };
      }

      const existingMessages = old.conversation.messages || [];

      // Handle messages with same ID (multiple content blocks)
      const existingIndex = existingMessages.findIndex(
        (m) => m.id === message.id,
      );

      if (existingIndex >= 0) {
        // Merge content blocks for messages with same ID
        const existing = existingMessages[existingIndex];
        const mergedContent = [
          ...(existing.content || []),
          ...(message.content || []),
        ];

        const result = {
          ...old,
          conversation: {
            ...old.conversation,
            messages: [
              ...existingMessages.slice(0, existingIndex),
              {
                ...existing,
                content: mergedContent,
                updatedOn: message.updatedOn,
              },
              ...existingMessages.slice(existingIndex + 1),
            ],
            updatedOn: new Date().toISOString(),
          },
        };
        return result;
      } else {
        // Add new message
        const result = {
          ...old,
          conversation: {
            ...old.conversation,
            messages: [...existingMessages, message],
            updatedOn: new Date().toISOString(),
          },
        };
        return result;
      }
    });
  }

  /**
   * Remove message from TanStack Query cache (for rollback)
   */
  private removeMessageFromCache(messageId: string): void {
    const cacheKey = getRuntimeServiceGetConversationQueryKey(
      this.instanceId,
      this.conversationId,
    );

    queryClient.setQueryData<V1GetConversationResponse>(cacheKey, (old) => {
      if (!old?.conversation) return old;

      return {
        ...old,
        conversation: {
          ...old.conversation,
          messages:
            old.conversation.messages?.filter((m) => m.id !== messageId) || [],
          updatedOn: new Date().toISOString(),
        },
      };
    });
  }

  /**
   * Handle incoming tool result by merging it with the corresponding tool call
   */
  private handleToolResult(toolResult: any): void {
    const cacheKey = getRuntimeServiceGetConversationQueryKey(
      this.instanceId,
      this.conversationId,
    );

    // Find and merge with existing tool call
    queryClient.setQueryData<V1GetConversationResponse>(cacheKey, (old) => {
      if (!old?.conversation?.messages) return old;

      const updatedMessages = old.conversation.messages.map((msg) => ({
        ...msg,
        content: msg.content?.map((block) => {
          if (block.toolCall?.id === toolResult.id) {
            return { ...block, toolResult };
          }
          return block;
        }),
      }));

      return {
        ...old,
        conversation: {
          ...old.conversation,
          messages: updatedMessages,
          updatedOn: new Date().toISOString(),
        },
      };
    });
  }

  // ----- Error Handling -----

  /**
   * Handle streaming errors with rollback and user feedback
   */
  private handleStreamingError(error: any, userMessage: V1Message): void {
    // Set error state for UI display
    this.streamError.set(this.getDescriptiveError(error));

    // Roll back optimistic updates
    this.removeMessageFromCache(userMessage.id!);

    // Restore draft message so user can easily retry
    const textContent = userMessage.content?.[0]?.text || "";
    this.draftMessage.set(textContent);
  }

  /**
   * Get descriptive error message for user feedback
   */
  private getDescriptiveError(error: any): string {
    if (error.name === "AbortError") {
      return "Message sending was cancelled";
    }

    if (error.status >= 500) {
      return "Server is temporarily unavailable. Please try sending your message again.";
    }

    if (error.name === "NetworkError" || !navigator.onLine) {
      return "Connection lost. Check your internet connection and try again.";
    }

    if (error.status === 429) {
      return "Too many requests. Please wait a moment before trying again.";
    }

    // Generic error with retry guidance
    return "Failed to send message. Please try again or refresh the page.";
  }
}
