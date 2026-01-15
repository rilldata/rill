/**
 * Local Conversation Manager - Wraps web-common's ConversationManager with local URL selector
 */

import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceListConversationsQueryOptions,
  type RpcStatus,
  type V1ListConversationsResponse,
} from "@rilldata/web-common/runtime-client";
import { createQuery, type CreateQueryResult } from "@tanstack/svelte-query";
import { derived, get, type Readable } from "svelte/store";
import { Conversation } from "@rilldata/web-common/features/chat/core/conversation";
import { invalidateConversationsList, NEW_CONVERSATION_ID } from "@rilldata/web-common/features/chat/core/utils";
import { LocalURLConversationSelector } from "./local-conversation-selector";

/**
 * Local version of ConversationManager that uses /ai URLs instead of /{org}/{project}/-/ai URLs
 */
export class LocalConversationManager {
  private static readonly MAX_CONCURRENT_STREAMS = 3;

  private newConversation: Conversation;
  private newConversationUnsub: (() => void) | null = null;
  private conversations = new Map<string, Conversation>();
  private conversationSelector: LocalURLConversationSelector;
  private readonly agent?: string;

  constructor(
    public readonly instanceId: string,
    agent?: string,
  ) {
    this.agent = agent;
    this.createNewConversation();
    this.conversationSelector = new LocalURLConversationSelector();
  }

  public listConversationsQuery(): CreateQueryResult<
    V1ListConversationsResponse,
    RpcStatus
  > {
    return createQuery(
      getRuntimeServiceListConversationsQueryOptions(this.instanceId, {
        userAgentPattern: "rill%",
      }),
      queryClient,
    );
  }

  public getCurrentConversation(): Readable<Conversation> {
    return derived(
      [this.conversationSelector.currentConversationId],
      ([$conversationId]) => {
        if ($conversationId === NEW_CONVERSATION_ID) {
          return this.newConversation;
        }

        const existing = this.conversations.get($conversationId);
        if (existing) {
          return existing;
        }

        const conversation = new Conversation(
          this.instanceId,
          $conversationId,
          this.agent,
        );
        conversation.on("stream-start", () =>
          this.enforceMaxConcurrentStreams(),
        );
        this.conversations.set($conversationId, conversation);
        return conversation;
      },
    );
  }

  public selectConversation(conversationId: string): void {
    this.conversationSelector.selectConversation(conversationId);
  }

  public enterNewConversationMode(): void {
    this.conversationSelector.clearSelection();
  }

  public cleanup(): void {
    this.conversations.forEach((conversation) => {
      conversation.cleanup();
    });
    this.conversations.clear();
    this.newConversation.cleanup();
  }

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

  private getActiveStreamingConversations(): Conversation[] {
    return Array.from(this.conversations.values()).filter((conv) =>
      get(conv.isStreaming),
    );
  }

  private enforceMaxConcurrentStreams(): void {
    try {
      const streamingConversations = this.getActiveStreamingConversations();

      if (
        streamingConversations.length >=
        LocalConversationManager.MAX_CONCURRENT_STREAMS
      ) {
        const conversationsToStop = streamingConversations.slice(
          0,
          streamingConversations.length -
            LocalConversationManager.MAX_CONCURRENT_STREAMS +
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

  private handleConversationCreated(conversationId: string): void {
    this.rotateNewConversation(conversationId);
    this.conversationSelector.selectConversation(conversationId);
    void invalidateConversationsList(this.instanceId);
  }

  private rotateNewConversation(conversationId: string): void {
    this.conversations.set(conversationId, this.newConversation);
    this.createNewConversation();
  }
}

// Singleton management
const localConversationManagerInstances = new Map<string, LocalConversationManager>();

export function getLocalConversationManager(
  instanceId: string,
  agent?: string,
): LocalConversationManager {
  const key = `${instanceId}:${agent || "default"}`;
  if (!localConversationManagerInstances.has(key)) {
    localConversationManagerInstances.set(
      key,
      new LocalConversationManager(instanceId, agent),
    );
  }
  return localConversationManagerInstances.get(key)!;
}

export function cleanupLocalConversationManager(instanceId: string): void {
  const keysToDelete: string[] = [];
  for (const key of localConversationManagerInstances.keys()) {
    if (key.startsWith(`${instanceId}:`)) {
      const manager = localConversationManagerInstances.get(key);
      if (manager) {
        manager.cleanup();
      }
      keysToDelete.push(key);
    }
  }
  keysToDelete.forEach((key) => localConversationManagerInstances.delete(key));
}
