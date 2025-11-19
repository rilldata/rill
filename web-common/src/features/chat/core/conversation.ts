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
import {
  SSEFetchClient,
  SSEHttpError,
  type SSEMessage,
} from "@rilldata/web-common/runtime-client/sse-fetch-client";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, writable, type Readable } from "svelte/store";
import type { HTTPError } from "../../../runtime-client/fetchWrapper";
import { MessageContentType, MessageType, ToolName } from "./types";
import { getOptimisticMessageId, NEW_CONVERSATION_ID } from "./utils";

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

  // Private state
  private sseClient: SSEFetchClient | null = null;
  private hasReceivedFirstMessage = false;

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

    const prompt = get(this.draftMessage).trim();
    if (!prompt) throw new Error("Cannot send empty message");

    // Optimistic updates
    this.draftMessage.set("");
    this.streamError.set(null);
    this.isStreaming.set(true);
    this.hasReceivedFirstMessage = false;

    const userMessage = this.addOptimisticUserMessage(prompt);

    try {
      // Start streaming - this establishes the connection
      const streamPromise = this.startStreaming(prompt);

      // Wait for streaming to complete
      await streamPromise;

      // Stream has completed successfully
      options?.onStreamComplete?.(this.conversationId);
    } catch (error) {
      // Transport errors can occur at two different stages:
      // 1. Before streaming starts: message not persisted, needs rollback
      // 2. During streaming: message already persisted, no rollback needed
      console.error("[Conversation] Message send error:", {
        error,
        conversationId: this.conversationId,
        hasReceivedFirstMessage: this.hasReceivedFirstMessage,
      });
      this.handleTransportError(
        error,
        userMessage,
        this.hasReceivedFirstMessage,
      );
      options?.onError?.(this.formatTransportError(error));
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
      this.sseClient = new SSEFetchClient();

      // Set up SSE event handlers
      this.sseClient.on("message", (message: SSEMessage) => {
        // Mark that we've received data
        // Since server always emits user message first (after persisting),
        // receiving any message means the server has persisted our message
        this.hasReceivedFirstMessage = true;

        // Handle application-level errors sent via SSE
        if (message.type === "error") {
          this.handleServerError(message.data);
          return;
        }

        // Handle normal streaming data
        try {
          const response: V1CompleteStreamingResponse = JSON.parse(
            message.data,
          );
          this.processStreamingResponse(response);
        } catch (error) {
          console.error("Failed to parse streaming response:", error);
          this.streamError.set("Failed to process server response");
        }
      });

      this.sseClient.on("error", (error) => {
        // Transport errors only: connection, network, HTTP failures
        console.error("[SSE] Transport error:", {
          message: error.message,
          status: error instanceof SSEHttpError ? error.status : undefined,
          statusText:
            error instanceof SSEHttpError ? error.statusText : undefined,
          name: error.name,
        });
        this.streamError.set(this.formatTransportError(error));
      });

      this.sseClient.on("close", () => {
        // Stream closed - completion handled in sendMessage
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
    await this.sseClient.start(baseUrl, {
      method: "POST",
      body: requestBody,
    });
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
      // Skip ALL user messages from the stream
      // Server echoes back the user message
      // We've already added it optimistically, so we don't want duplicates
      // Note: Server generates new IDs for streamed messages, can't match by ID
      if (response.message.role === "user") {
        return;
      }

      this.addMessageToCache(response.message);
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
      // Transfer the conversation data and messages to the real conversation ID cache
      queryClient.setQueryData<V1GetConversationResponse>(newCacheKey, {
        conversation: {
          ...existingData.conversation,
          id: realConversationId,
        },
        messages: existingData.messages || [],
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
      type: MessageType.CALL,
      tool: ToolName.ROUTER_AGENT,
      contentType: MessageContentType.JSON,
      contentData: JSON.stringify({ prompt }),
      createdOn: new Date().toISOString(),
      updatedOn: new Date().toISOString(),
    };

    this.addMessageToCache(userMessage);
    return userMessage;
  }

  /**
   * Add message to TanStack Query cache
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
            createdOn: message.createdOn,
            updatedOn: new Date().toISOString(),
          },
          messages: [message],
        };
      }

      const existingMessages = old.messages || [];

      // Add new message to the end of the list
      return {
        ...old,
        conversation: {
          ...old.conversation,
          updatedOn: new Date().toISOString(),
        },
        messages: [...existingMessages, message],
      };
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
          updatedOn: new Date().toISOString(),
        },
        messages: old.messages?.filter((m) => m.id !== messageId) || [],
      };
    });
  }

  // ----- Error Handling -----
  // Error handling is split by error type (server vs transport) and responsibility
  // (formatting vs handling). Each error type has a formatter (pure) and handler (side effects).

  // ----- Server Errors (Application-level) -----

  /**
   * Format server error data into user-friendly message
   */
  private formatServerError(errorData: string): string {
    try {
      const parsed = JSON.parse(errorData);
      return parsed.error || "Server error occurred";
    } catch {
      return `Server error: ${errorData}`;
    }
  }

  /**
   * Handle server-sent errors (event: error from SSE)
   *
   * These are application-level errors (AI failures, tool errors, etc.) that occur
   * AFTER the server has already persisted the user's message. No rollback is needed -
   * the user's message should remain visible in the conversation with an error indicator.
   */
  private handleServerError(errorData: string): void {
    this.streamError.set(this.formatServerError(errorData));
  }

  // ----- Transport Errors (Connection-level) -----

  /**
   * Format transport error into user-friendly message
   */
  private formatTransportError(error: Error): string {
    if (error.name === "AbortError") {
      return "Message sending was cancelled";
    }

    // Extract status code from SSEHttpError
    const status = error instanceof SSEHttpError ? error.status : null;

    // Authentication errors - suggest refresh to get new JWT
    if (status === 401 || status === 403) {
      return "Authentication failed. Please refresh the page and try again.";
    }

    // Bad request errors
    if (status === 400) {
      return "Invalid request. Please try again.";
    }

    // Server errors (5xx)
    if (status && status >= 500 && status < 600) {
      return "Server is temporarily unavailable. Please try sending your message again.";
    }

    // Rate limiting
    if (status === 429) {
      return "Too many requests. Please wait a moment before trying again.";
    }

    // Network/connection errors (fetch() throws TypeError for network failures)
    const lowerMessage = error.message?.toLowerCase() || "";
    const isNetworkError =
      (error.name === "TypeError" &&
        (lowerMessage.includes("fetch") ||
          lowerMessage.includes("network") ||
          lowerMessage.includes("load failed"))) ||
      (typeof navigator !== "undefined" && !navigator.onLine);

    if (isNetworkError) {
      return "Unable to connect to server. Please check your connection and try again.";
    }

    // Fallback error message
    return "Failed to connect to server. Please try again or refresh the page.";
  }

  /**
   * Handle transport-level errors with conditional rollback
   *
   * Transport errors can occur at two stages:
   * 1. Before streaming starts: Connection failures, HTTP errors before request completes
   *    → Rollback needed (message never reached server)
   * 2. During streaming: Network drops, server crashes while streaming responses
   *    → No rollback (message already persisted on server)
   */
  private handleTransportError(
    error: any,
    userMessage: V1Message,
    wasStreaming: boolean,
  ): void {
    // Set error message
    this.streamError.set(this.formatTransportError(error));

    // Only rollback if we hadn't started streaming yet
    if (!wasStreaming) {
      // Message never reached server - remove optimistic update
      this.removeMessageFromCache(userMessage.id!);

      // Restore draft message so user can easily retry
      const textContent = userMessage.contentData || "";
      this.draftMessage.set(textContent);
    }
    // If we were streaming, message is already on server - keep it in UI
  }
}
