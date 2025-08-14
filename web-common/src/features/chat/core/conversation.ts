import { page } from "$app/stores";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetConversationQueryKey,
  getRuntimeServiceGetConversationQueryOptions,
  getRuntimeServiceListConversationsQueryKey,
  runtimeServiceComplete,
  type RpcStatus,
  type V1CompleteResponse,
  type V1GetConversationResponse,
  type V1ListConversationsResponse,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, writable, type Readable } from "svelte/store";
import {
  detectAppContext,
  formatChatError,
  getOptimisticMessageId,
  NEW_CONVERSATION_ID,
} from "./chat-utils";

/**
 * Individual conversation state management.
 *
 * Handles message sending, optimistic updates, and conversation-specific queries
 * for a single conversation.
 */
export class Conversation {
  public readonly draftMessage = writable<string>("");
  public readonly isSendingMessage = writable(false);
  public readonly errorMessage = writable<string | null>(null);

  constructor(
    private readonly instanceId: string,
    public readonly conversationId: string,
    private readonly options?: {
      onConversationCreated?: (conversationId: string) => void;
    },
  ) {}

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
            enabled: this.conversationId !== NEW_CONVERSATION_ID,
          },
        },
      ),
      queryClient,
    );
  }

  get canSendMessage(): Readable<boolean> {
    return derived([this.isSendingMessage], ([$isSending]) => !$isSending);
  }

  // ACTIONS

  async sendMessage(options?: {
    onSuccess?: (response: V1CompleteResponse) => void;
    onError?: (failedMessage: string) => void;
  }): Promise<V1CompleteResponse> {
    if (this.conversationId === NEW_CONVERSATION_ID) {
      return await this.sendInitialMessage(options);
    } else {
      return await this.sendSubsequentMessage(options);
    }
  }

  private async sendInitialMessage(options?: {
    onSuccess?: (response: V1CompleteResponse) => void;
    onError?: (failedMessage: string) => void;
  }): Promise<V1CompleteResponse> {
    const initialMessage = get(this.draftMessage);

    if (!initialMessage.trim()) {
      throw new Error("Cannot start a conversation with an empty message");
    }

    // 1. Optimistic UI updates
    this.draftMessage.set("");
    this.errorMessage.set(null);
    this.isSendingMessage.set(true);

    const getNewConversationQueryKey = getRuntimeServiceGetConversationQueryKey(
      this.instanceId,
      this.conversationId,
    );

    // 2. Cancel outgoing refetches and snapshot previous value
    await queryClient.cancelQueries({
      queryKey: getNewConversationQueryKey,
    });

    // 3. Optimistically add user message to the `GetConversation` query
    const userMessage: V1Message = {
      id: getOptimisticMessageId(),
      role: "user" as const,
      content: [{ text: initialMessage }],
      createdOn: new Date().toISOString(),
      updatedOn: new Date().toISOString(),
    };

    queryClient.setQueryData<V1GetConversationResponse>(
      getNewConversationQueryKey,
      () => {
        return {
          conversation: {
            messages: [userMessage],
            updatedOn: new Date().toISOString(),
          },
        };
      },
    );

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

      options?.onSuccess?.(response);

      // Call the conversation created callback to promote this to a real conversation
      this.options?.onConversationCreated?.(realConversationId);

      // Reset the `GetNewConversation` query cache
      queryClient.removeQueries({ queryKey: getNewConversationQueryKey });

      return response;
    } catch (error) {
      console.error("Failed to start new conversation:", error);
      throw error;
    }
  }

  private async sendSubsequentMessage(options?: {
    onSuccess?: (response: V1CompleteResponse) => void;
    onError?: (failedMessage: string) => void;
  }): Promise<V1CompleteResponse> {
    const text = get(this.draftMessage);
    if (!text.trim()) {
      return Promise.reject(new Error("Cannot send an empty message."));
    }

    // 1. Optimistic UI updates
    this.draftMessage.set("");
    this.errorMessage.set(null);
    this.isSendingMessage.set(true);

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
      this.isSendingMessage.set(false);
      queryClient.invalidateQueries({ queryKey: cacheKey });
      options?.onSuccess?.(response);

      return response;
    } catch (error) {
      // 6. On error, roll back the optimistic update
      if (previousConversation) {
        queryClient.setQueryData(cacheKey, previousConversation);
      }

      const errorMessage = formatChatError(error as Error);
      this.errorMessage.set(errorMessage);
      this.isSendingMessage.set(false);

      // Fire the error callback
      options?.onError?.(text);

      // Re-throw the error so the caller can handle it
      throw new Error(errorMessage);
    }
  }
}
