import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetConversationQueryKey,
  getRuntimeServiceGetConversationQueryOptions,
  runtimeServiceForkConversation,
  type RpcStatus,
  type RuntimeServiceCompleteBody,
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
import {
  derived,
  get,
  writable,
  type Readable,
  type Writable,
} from "svelte/store";
import type { HTTPError } from "../../../runtime-client/fetchWrapper";
import { transformToBlocks, type Block } from "./messages/block-transform";
import { MessageContentType, MessageType, ToolName } from "./types";
import {
  getOptimisticMessageId,
  invalidateConversationsList,
  NEW_CONVERSATION_ID,
} from "./utils";
import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";
import { getToolConfig } from "@rilldata/web-common/features/chat/core/messages/tools/tool-registry.ts";

type ConversationEvents = {
  "conversation-created": string;
  "conversation-forked": string;
  "stream-start": void;
  message: V1Message;
  "stream-complete": string;
  error: string;
};

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

  private readonly events = new EventEmitter<ConversationEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;
  public readonly once = this.events.once.bind(
    this.events,
  ) as typeof this.events.once;

  // Private state
  private sseClient: SSEFetchClient | null = null;
  private hasReceivedFirstMessage = false;

  // Reactive conversation ID - enables query to auto-update when ID changes (e.g., after fork)
  private readonly conversationIdStore: Writable<string>;
  private readonly conversationQuery: CreateQueryResult<
    V1GetConversationResponse,
    RpcStatus
  >;

  public get conversationId(): string {
    return get(this.conversationIdStore);
  }

  private set conversationId(value: string) {
    this.conversationIdStore.set(value);
  }

  constructor(
    private readonly instanceId: string,
    initialConversationId: string,
    private readonly agent: string = "", // Empty string lets the router agent decide
  ) {
    this.conversationIdStore = writable(initialConversationId);

    // Create query with reactive options that respond to conversationId changes
    const queryOptionsStore = derived(
      this.conversationIdStore,
      ($conversationId) =>
        getRuntimeServiceGetConversationQueryOptions(
          this.instanceId,
          $conversationId,
          {
            query: {
              enabled: $conversationId !== NEW_CONVERSATION_ID,
              staleTime: Infinity, // We manage cache manually during streaming
            },
          },
        ),
    );
    this.conversationQuery = createQuery(queryOptionsStore, queryClient);
  }

  /**
   * Get ownership status from the conversation query.
   * Returns true if the current user owns this conversation or if ownership is unknown.
   */
  private getIsOwner(): boolean {
    if (this.conversationId === NEW_CONVERSATION_ID) {
      return true; // New conversations are always owned by the creator
    }

    // Default to true if query hasn't loaded yet (optimistic assumption)
    return get(this.conversationQuery).data?.isOwner ?? true;
  }

  // ===== PUBLIC API =====

  /**
   * Get a reactive query for this conversation's data.
   * The query reacts to conversationId changes (e.g., after fork).
   */
  public getConversationQuery(): CreateQueryResult<
    V1GetConversationResponse,
    RpcStatus
  > {
    return this.conversationQuery;
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
   * Get message blocks for rendering this conversation.
   * Transforms raw messages into a structured list of blocks (text, thinking blocks, charts).
   */
  public getBlocks(): Readable<Block[]> {
    return derived(
      [this.getConversationQuery(), this.isStreaming],
      ([$query, $isStreaming]) => {
        const messages = $query.data?.messages ?? [];
        const isLoading = !!$query.isLoading;

        return transformToBlocks(messages, $isStreaming, isLoading);
      },
    );
  }

  /**
   * Send a message and handle streaming response
   *
   * @param context - Chat context to be sent with the message
   * @param options - Callback functions for different stages of message sending
   */
  public async sendMessage(
    context: RuntimeServiceCompleteBody,
    options?: { onStreamStart?: () => void },
  ): Promise<void> {
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

    // Fork conversation if user is not the owner (viewing a shared conversation)
    const isOwner = this.getIsOwner();
    if (!isOwner && this.conversationId !== NEW_CONVERSATION_ID) {
      try {
        const forkedConversationId = await this.forkConversation();
        // Update to the forked conversation (setter updates the reactive store)
        this.conversationId = forkedConversationId;
        this.events.emit("conversation-forked", forkedConversationId);
      } catch (error) {
        console.error("[Conversation] Fork failed:", error);
        this.isStreaming.set(false);
        this.streamError.set(
          "Failed to create your copy of this conversation. Please try again.",
        );
        return;
      }
    }

    const userMessage = this.addOptimisticUserMessage(prompt);

    try {
      options?.onStreamStart?.(); // Callback for direct callers
      this.events.emit("stream-start"); // Event for external listeners
      // Start streaming - this establishes the connection
      const streamPromise = this.startStreaming(prompt, context);

      // Wait for streaming to complete
      await streamPromise;

      // Stream has completed successfully
      this.events.emit("stream-complete", this.conversationId);

      // Temporary fix to make sure the title of the conversation is updated.
      void invalidateConversationsList(this.instanceId);
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
      this.events.emit("error", this.formatTransportError(error));
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

    this.events.clearListeners();
  }

  // ===== PRIVATE IMPLEMENTATION =====

  // ----- Transport Layer: SSE Connection Management -----

  /**
   * Start streaming completion responses for a given prompt
   * Returns a Promise that resolves when streaming completes
   */
  private async startStreaming(
    prompt: string,
    context: RuntimeServiceCompleteBody | undefined,
  ): Promise<void> {
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
          if (response.message) this.events.emit("message", response.message);
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
      agent: this.agent,
      ...context,
    };

    // Notify that streaming is about to start (for concurrent stream management)
    this.events.emit("stream-start");

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
      if (response.message.type === MessageType.CALL) {
        const config = getToolConfig(response.message.tool);
        config?.onResult?.(response.message);
      }
    }
  }

  // ----- Conversation Lifecycle -----

  /**
   * Fork the current conversation to create a copy owned by the current user.
   * Used when a non-owner wants to continue a shared conversation.
   *
   * Note: The cache copying logic here follows the pattern established by
   * `transitionToRealConversation`—both read from an old cache key and write
   * to a new one with an updated conversation ID. However, since forking
   * conceptually creates a new conversation from an existing one, this
   * responsibility might be better suited for ConversationManager in the future.
   */
  private async forkConversation(): Promise<string> {
    const originalConversationId = this.conversationId;

    const response = await runtimeServiceForkConversation(
      this.instanceId,
      this.conversationId,
      {},
    );

    if (!response.conversationId) {
      throw new Error("Fork response missing conversation ID");
    }

    const forkedConversationId = response.conversationId;

    // Copy cached messages from original conversation to forked conversation
    // This ensures the UI shows the conversation history immediately
    const originalCacheKey = getRuntimeServiceGetConversationQueryKey(
      this.instanceId,
      originalConversationId,
    );
    const forkedCacheKey = getRuntimeServiceGetConversationQueryKey(
      this.instanceId,
      forkedConversationId,
    );
    const originalData =
      queryClient.getQueryData<V1GetConversationResponse>(originalCacheKey);

    if (originalData) {
      queryClient.setQueryData<V1GetConversationResponse>(forkedCacheKey, {
        conversation: {
          ...originalData.conversation,
          id: forkedConversationId,
        },
        messages: originalData.messages || [],
        isOwner: true, // User now owns the forked conversation
      });
    }

    // Invalidate the conversations list to show the new forked conversation
    void invalidateConversationsList(this.instanceId);

    return forkedConversationId;
  }

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

    // Update the conversation ID (setter updates the reactive store)
    this.conversationId = realConversationId;

    // Notify that conversation was created
    this.events.emit("conversation-created", realConversationId);
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
