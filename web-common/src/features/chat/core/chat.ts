/**
 * Main composable for chat functionality - use this in components.
 *
 * Provides a high-level API that coordinates queries, mutations, and state
 * for chat features. This is the primary interface components should use.
 */
import { page } from "$app/stores";
import type { Page } from "@sveltejs/kit";
import { createQuery } from "@tanstack/svelte-query";
import { derived, get, writable } from "svelte/store";
import { queryClient } from "../../../lib/svelte-query/globalQueryClient";
import type {
  V1AppContext,
  V1Conversation,
  V1GetConversationResponse,
  V1Message,
} from "../../../runtime-client";
import {
  createRuntimeServiceComplete,
  getRuntimeServiceGetConversationQueryKey,
  getRuntimeServiceGetConversationQueryOptions,
  getRuntimeServiceListConversationsQueryKey,
  getRuntimeServiceListConversationsQueryOptions,
  V1AppContextType,
} from "../../../runtime-client";
import { runtime } from "../../../runtime-client/runtime-store";

// =============================================================================
// INTERNAL CONSTANTS
// =============================================================================

// Conversation & Message ID Prefixes
const OPTIMISTIC_CONVERSATION_ID_PREFIX = "optimistic-conversation-";
const OPTIMISTIC_MESSAGE_ID_PREFIX = "optimistic-message-";

// =============================================================================
// INTERNAL HELPER FUNCTIONS
// =============================================================================

function getOptimisticConversationId(): string {
  return `${OPTIMISTIC_CONVERSATION_ID_PREFIX}${Date.now()}`;
}

function getOptimisticMessageId(): string {
  return `${OPTIMISTIC_MESSAGE_ID_PREFIX}${Date.now()}`;
}

export function isOptimisticConversationId(id: string): boolean {
  return id.startsWith(OPTIMISTIC_CONVERSATION_ID_PREFIX);
}

// Cache query keys
function getConversationCacheKey(instanceId: string, conversationId: string) {
  if (isOptimisticConversationId(conversationId)) {
    return ["conversation", instanceId, "optimistic", conversationId];
  } else {
    return getRuntimeServiceGetConversationQueryKey(instanceId, conversationId);
  }
}

// App context detection (determines what resources/context the chat can see)
function detectAppContext(page: Page): V1AppContext | null {
  const routeId = page.route.id;

  switch (routeId) {
    case "/[organization]/[project]/-/chat":
    case "/[organization]/[project]/-/chat/[conversationId]":
      return {
        contextType: V1AppContextType.APP_CONTEXT_TYPE_PROJECT_CHAT,
      };
    case "/[organization]/[project]/explore/[dashboard]":
      return {
        contextType: V1AppContextType.APP_CONTEXT_TYPE_EXPLORE_DASHBOARD,
        contextMetadata: {
          dashboard_name: page.params.dashboard,
        },
      };
    case "/(viz)/explore/[name]":
      return {
        contextType: V1AppContextType.APP_CONTEXT_TYPE_EXPLORE_DASHBOARD,
        contextMetadata: {
          dashboard_name: page.params.name,
        },
      };
    default:
      return null;
  }
}

// =============================================================================
// PUBLIC API
// =============================================================================

export interface ChatCoreOptions {
  initialConversationId?: string | null;
  onConversationChange?: (conversationId: string | null) => void;
}

export function useChatCore(options?: ChatCoreOptions) {
  // Instance-specific conversation state
  const currentConversationId = writable<string | null>(
    options?.initialConversationId ?? null,
  );

  // Notify parent when conversation changes
  if (options?.onConversationChange) {
    currentConversationId.subscribe(options.onConversationChange);
  }

  // Messages derived from TanStack Query cache (our single source of truth)
  const messages = derived(
    [runtime, currentConversationId],
    ([$runtime, $currentConversationId], set) => {
      if (!$runtime?.instanceId || !$currentConversationId) {
        set([]);
        return;
      }

      // Get messages from TanStack Query cache (works for both optimistic and committed conversations)
      const queryKey = getConversationCacheKey(
        $runtime.instanceId,
        $currentConversationId,
      );

      const cachedData = queryClient.getQueryData(queryKey) as
        | V1GetConversationResponse
        | undefined;
      const messages = cachedData?.conversation?.messages || [];
      set(messages);

      // Subscribe to cache changes
      const unsubscribe = queryClient.getQueryCache().subscribe((event) => {
        if (event.query.queryKey.toString() === queryKey.toString()) {
          const updatedData = event.query.state.data as
            | V1GetConversationResponse
            | undefined;
          const updatedMessages = updatedData?.conversation?.messages || [];
          set(updatedMessages);
        }
      });

      return unsubscribe;
    },
    [] as V1Message[],
  );

  // =============================================================================
  // SHARED GLOBAL MUTATIONS
  // =============================================================================

  // Mutation for creating a new conversation with the first message
  const completeNewConversation = createRuntimeServiceComplete(
    {
      mutation: {
        onMutate: async (variables) => {
          const { instanceId, data } = variables;

          // Create optimistic conversation ID and cache key
          const optimisticConversationId = getOptimisticConversationId();
          const optimisticCacheKey = getConversationCacheKey(
            instanceId,
            optimisticConversationId,
          );

          // Create user message from the mutation data
          const userMessage: V1Message = {
            id: getOptimisticMessageId(),
            role: data.messages?.[0]?.role || "user",
            content: data.messages?.[0]?.content || [],
            createdOn: new Date().toISOString(),
            updatedOn: new Date().toISOString(),
          };

          // Create optimistic conversation in cache
          queryClient.setQueryData(
            optimisticCacheKey,
            (): V1GetConversationResponse => ({
              conversation: {
                id: optimisticConversationId,
                createdOn: new Date().toISOString(),
                updatedOn: new Date().toISOString(),
                messages: [userMessage],
              },
            }),
          );

          // Update current conversation ID to optimistic ID
          currentConversationId.set(optimisticConversationId);

          return { optimisticConversationId };
        },

        onSuccess: (response, variables, context) => {
          const { instanceId } = variables;

          // Invalidate conversation list to show new conversation
          void queryClient.invalidateQueries({
            queryKey: getRuntimeServiceListConversationsQueryKey(instanceId),
          });

          if (response.conversationId && context?.optimisticConversationId) {
            // Create cache for the new committed conversation
            queryClient.setQueryData(
              getRuntimeServiceGetConversationQueryKey(
                instanceId,
                response.conversationId,
              ),
              (): V1GetConversationResponse => ({
                conversation: {
                  id: response.conversationId!,
                  title: "Loading...", // Will be updated by list invalidation
                  createdOn: new Date().toISOString(),
                  updatedOn: new Date().toISOString(),
                  messages: response.messages || [],
                },
              }),
            );

            // Remove optimistic cache entry
            const optimisticCacheKey = getConversationCacheKey(
              instanceId,
              context.optimisticConversationId,
            );
            void queryClient.removeQueries({ queryKey: optimisticCacheKey });

            // Update current conversation ID to committed ID
            currentConversationId.set(response.conversationId);
          }
        },

        onError: (err, variables, context) => {
          const { instanceId } = variables;

          // Clean up optimistic conversation on error
          if (context?.optimisticConversationId) {
            const optimisticCacheKey = getConversationCacheKey(
              instanceId,
              context.optimisticConversationId,
            );
            void queryClient.removeQueries({ queryKey: optimisticCacheKey });
          }

          // Reset current conversation ID to null
          currentConversationId.set(null);
          console.error("Chat: Failed to create new conversation", err);
        },
      },
    },
    queryClient,
  );

  // Mutation for adding a message to an existing conversation
  const completeExistingConversation = createRuntimeServiceComplete(
    {
      mutation: {
        onMutate: async (variables) => {
          const { instanceId, data } = variables;
          const conversationId = data.conversationId!; // Should always exist for this mutation

          const cacheKey = getConversationCacheKey(instanceId, conversationId);
          await queryClient.cancelQueries({ queryKey: cacheKey });

          // Snapshot previous state for rollback
          const previousConversation = queryClient.getQueryData(cacheKey);

          // Add optimistic user message to existing conversation
          if (data.messages && data.messages[0]) {
            queryClient.setQueryData(
              cacheKey,
              (
                old: V1GetConversationResponse | undefined,
              ): V1GetConversationResponse => {
                if (!old?.conversation) {
                  return old as V1GetConversationResponse;
                }

                const userMessage = data.messages![0];
                const optimisticMessage: V1Message = {
                  id: getOptimisticMessageId(),
                  role: userMessage.role,
                  content: userMessage.content,
                  createdOn: new Date().toISOString(),
                  updatedOn: new Date().toISOString(),
                };

                return {
                  conversation: {
                    ...old.conversation,
                    messages: [
                      ...(old.conversation.messages || []),
                      optimisticMessage,
                    ],
                    updatedOn: new Date().toISOString(),
                  },
                };
              },
            );
          }

          return { previousConversation, conversationId };
        },

        onSuccess: (response, variables) => {
          const { instanceId } = variables;

          if (response.conversationId) {
            const queryKey = getRuntimeServiceGetConversationQueryKey(
              instanceId,
              response.conversationId,
            );

            const oldData = queryClient.getQueryData(queryKey) as
              | V1GetConversationResponse
              | undefined;
            const existingMessages = oldData?.conversation?.messages || [];
            const newMessages = response.messages || [];

            // Merge: keep existing messages, add new ones that don't already exist
            const existingMessageIds = new Set(
              existingMessages.map((msg) => msg.id).filter(Boolean),
            );
            const messagesToAdd = newMessages.filter(
              (msg) => msg.id && !existingMessageIds.has(msg.id),
            );

            // Combine all messages
            const allMessages = [...existingMessages, ...messagesToAdd];

            // Remove optimistic messages if a committed message with same role and content exists
            const finalMessages = allMessages.filter((msg) => {
              if (msg.id && msg.id.startsWith(OPTIMISTIC_MESSAGE_ID_PREFIX)) {
                // Check if a committed message exists with same role and content
                return !allMessages.some(
                  (other) =>
                    other !== msg &&
                    other.role === msg.role &&
                    JSON.stringify(other.content) ===
                      JSON.stringify(msg.content) &&
                    other.id &&
                    !other.id.startsWith(OPTIMISTIC_MESSAGE_ID_PREFIX),
                );
              }
              return true;
            });

            queryClient.setQueryData(
              queryKey,
              (
                old: V1GetConversationResponse | undefined,
              ): V1GetConversationResponse => ({
                conversation: {
                  ...(old?.conversation || {}),
                  id: response.conversationId!,
                  title: old?.conversation?.title || "Loading...",
                  createdOn:
                    old?.conversation?.createdOn || new Date().toISOString(),
                  updatedOn: new Date().toISOString(),
                  messages: finalMessages,
                },
              }),
            );
          }
        },

        onError: (err, variables, context) => {
          const { instanceId, data } = variables;
          const conversationId = data.conversationId!;

          // Rollback optimistic update
          if (context?.previousConversation) {
            const cacheKey = getConversationCacheKey(
              instanceId,
              conversationId,
            );
            queryClient.setQueryData(cacheKey, context.previousConversation);
          }
          console.error(
            "Chat: Failed to add message to existing conversation",
            err,
          );
        },
      },
    },
    queryClient,
  );

  // Global derived state (shared across all components)
  const loading = derived(
    [completeNewConversation, completeExistingConversation],
    ([$newConversation, $existingConversation]) =>
      $newConversation.isPending || $existingConversation.isPending,
    false,
  );

  const error = derived(
    [completeNewConversation, completeExistingConversation],
    ([$newConversation, $existingConversation]) => {
      // Check both mutations for errors
      const newError = $newConversation.isError ? $newConversation.error : null;
      const existingError = $existingConversation.isError
        ? $existingConversation.error
        : null;
      const error = newError || existingError;

      if (!error) {
        return null;
      }

      // Format error messages for better UX
      if (error instanceof Error) {
        if (
          error.message.includes("instance") ||
          error.message.includes("API client")
        ) {
          return error.message;
        } else if (
          error.message.includes("Network Error") ||
          error.message.includes("fetch")
        ) {
          return "Could not connect to the server. Please check your connection.";
        } else {
          return `Failed to send message: ${error.message}`;
        }
      } else {
        return "Could not get a response from the server.";
      }
    },
    null,
  );

  // Instance-specific actions
  const coreActions = {
    createNewConversation(): void {
      currentConversationId.set(null);
    },

    selectConversation(conv: V1Conversation): void {
      currentConversationId.set(conv.id || null);
    },

    async sendMessage(messageText: string): Promise<string | null> {
      if (!messageText.trim()) return null;

      const $runtime = get(runtime);
      const currentId = get(currentConversationId);

      // Detect context based on current page
      const appContext = detectAppContext(get(page));

      try {
        if (!currentId) {
          // No current conversation - create a new one
          const $completeNewConversationMutation = get(completeNewConversation);
          const response = await $completeNewConversationMutation.mutateAsync({
            instanceId: $runtime.instanceId,
            data: {
              // No conversationId for new conversations
              appContext: appContext || undefined,
              messages: [{ role: "user", content: [{ text: messageText }] }],
            },
          });
          return response.conversationId || null;
        } else {
          // Existing conversation - add message to it
          const $completeExistingConversationMutation = get(
            completeExistingConversation,
          );
          await $completeExistingConversationMutation.mutateAsync({
            instanceId: $runtime.instanceId,
            data: {
              conversationId: currentId,
              messages: [{ role: "user", content: [{ text: messageText }] }],
            },
          });
          return currentId;
        }
      } catch (e) {
        console.error("Failed to send message:", e);
        throw e;
      }
    },
  };

  // Queries
  const listConversationsQueryOptionsStore = derived([runtime], ([$runtime]) =>
    getRuntimeServiceListConversationsQueryOptions($runtime.instanceId, {
      query: {
        enabled: true,
      },
    }),
  );
  const listConversationsQuery = createQuery(
    listConversationsQueryOptionsStore,
  );

  const getConversationQueryOptionsStore = derived(
    [runtime, currentConversationId],
    ([$runtime, $currentConversationId]) =>
      getRuntimeServiceGetConversationQueryOptions(
        $runtime.instanceId,
        $currentConversationId || "",
        undefined,
        {
          query: {
            enabled:
              !!$currentConversationId &&
              !isOptimisticConversationId($currentConversationId),
            staleTime: 5 * 60 * 1000, // 5 minutes
            gcTime: 10 * 60 * 1000, // 10 minutes
          },
        },
      ),
  );
  const getConversationQuery = createQuery(getConversationQueryOptionsStore);

  // Derived state
  const listConversationsData = derived(
    listConversationsQuery,
    ($query) => $query.data,
  );
  const getConversationData = derived(
    getConversationQuery,
    ($query) => $query.data,
  );
  const currentConversation = derived(
    getConversationData,
    ($data) => $data?.conversation || null,
  );

  // Component-level action wrappers
  async function handleSendMessage(
    message: string,
    onSuccess?: (conversationId: string | null) => void,
    onError?: (message: string) => void,
  ) {
    try {
      const conversationId = await coreActions.sendMessage(message);
      onSuccess?.(conversationId);
    } catch (error) {
      console.error("Failed to send message:", error);
      onError?.(message);
    }
  }

  function createNewConversation(onSuccess?: () => void) {
    coreActions.createNewConversation();
    onSuccess?.();
  }

  function selectConversation(conv: V1Conversation, onSuccess?: () => void) {
    coreActions.selectConversation(conv);
    onSuccess?.();
  }

  return {
    // Queries
    listConversationsQuery,
    getConversationQuery,

    // Derived state
    listConversationsData,
    currentConversation,
    isConversationLoading: derived(
      getConversationQuery,
      ($query) => $query.isLoading,
    ),

    // Instance state
    currentConversationId,
    loading,
    error,
    messages,

    // Actions
    handleSendMessage,
    createNewConversation,
    selectConversation,
  };
}
