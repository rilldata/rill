import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetConversationQueryKey,
  getRuntimeServiceGetConversationQueryOptions,
  runtimeServiceComplete,
  type RpcStatus,
  type V1CompleteResponse,
  type V1GetConversationResponse,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, writable, type Readable } from "svelte/store";
import { formatChatError, getOptimisticMessageId } from "./chat-utils";

/**
 * Individual conversation state management.
 *
 * Handles message sending, optimistic updates, and conversation-specific queries
 * for a single conversation.
 */
export class Conversation {
  public readonly draftMessage = writable<string>("");
  public readonly isSending = writable(false);
  public readonly sendError = writable<string | null>(null);

  private constructor(
    private readonly instanceId: string,
    private readonly conversationId: string,
  ) {}

  /**
   * Create a Conversation instance for an existing conversation
   */
  static fromExistingId(
    instanceId: string,
    conversationId: string,
  ): Conversation {
    return new Conversation(instanceId, conversationId);
  }

  // QUERIES

  getConversationQuery(): CreateQueryResult<
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
            enabled: true,
          },
        },
      ),
      queryClient,
    );
  }

  get canSendMessage(): Readable<boolean> {
    return derived([this.isSending], ([$isSending]) => !$isSending);
  }

  // ACTIONS

  async sendMessage(options?: {
    onSuccess?: (response: V1CompleteResponse) => void;
    onError?: (failedMessage: string) => void;
  }): Promise<V1CompleteResponse> {
    const text = get(this.draftMessage);
    if (!text.trim()) {
      return Promise.reject(new Error("Cannot send an empty message."));
    }

    // 1. Optimistic UI updates
    this.draftMessage.set("");
    this.sendError.set(null);
    this.isSending.set(true);

    const cacheKey = getRuntimeServiceGetConversationQueryKey(
      this.instanceId,
      this.conversationId,
    );

    // 2. Cancel outgoing refetches and snapshot previous value
    await queryClient.cancelQueries({ queryKey: cacheKey });
    const previousConversation =
      queryClient.getQueryData<V1GetConversationResponse>(cacheKey);

    // 3. Optimistically add user message
    const userMessage: V1Message = {
      id: getOptimisticMessageId(),
      role: "user",
      content: [{ text }],
      createdOn: new Date().toISOString(),
      updatedOn: new Date().toISOString(),
    };

    // Update the existing conversation
    queryClient.setQueryData<V1GetConversationResponse>(cacheKey, (old) => {
      if (!old?.conversation) return old;
      return {
        ...old,
        conversation: {
          ...old.conversation,
          messages: [...(old.conversation.messages || []), userMessage],
          updatedOn: new Date().toISOString(),
        },
      };
    });

    try {
      // 4. Make the API call
      const response = await runtimeServiceComplete(this.instanceId, {
        conversationId: this.conversationId,
        messages: [
          {
            role: "user" as const,
            content: [{ text }],
          },
        ],
      });

      // 5. On success, clear draft, invalidate query, and fire callback
      this.isSending.set(false);
      queryClient.invalidateQueries({ queryKey: cacheKey });
      options?.onSuccess?.(response);

      return response;
    } catch (error) {
      // 6. On error, roll back the optimistic update
      if (previousConversation) {
        queryClient.setQueryData(cacheKey, previousConversation);
      }

      const errorMessage = formatChatError(error as Error);
      this.sendError.set(errorMessage);
      this.isSending.set(false);

      // Fire the error callback
      options?.onError?.(text);

      // Re-throw the error so the caller can handle it
      throw new Error(errorMessage);
    }
  }
}
