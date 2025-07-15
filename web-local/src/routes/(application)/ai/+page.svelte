<script lang="ts">
  import { onMount } from "svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    createRuntimeServiceComplete,
    createRuntimeServiceListConversations,
    createRuntimeServiceGetConversation,
  } from "@rilldata/web-common/runtime-client/gen/runtime-service/runtime-service";
  import type {
    V1Message,
    V1Conversation,
    RuntimeServiceCompleteBody,
  } from "@rilldata/web-common/runtime-client/gen/index.schemas";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { get } from "svelte/store";

  let messages: V1Message[] = [];
  let inputValue = "";
  let isLoading = false;
  let currentConversationId: string | null = null;
  let conversations: V1Conversation[] = [];
  let messagesContainer: HTMLDivElement;

  const instanceId = "default";

  // Enable chat feature flag for this page
  onMount(() => {
    featureFlags.set(true, "chat");
  });

  // API clients
  const completeMutation = createRuntimeServiceComplete();
  const listConversationsQuery =
    createRuntimeServiceListConversations(instanceId);

  let getConversationQuery: any = null;

  // Reactive statements
  $: if ($listConversationsQuery.data?.conversations) {
    conversations = $listConversationsQuery.data.conversations;
  }

  $: if (
    getConversationQuery &&
    $getConversationQuery.data?.conversation?.messages
  ) {
    messages = $getConversationQuery.data.conversation.messages;
    scrollToBottom();
  }

  function scrollToBottom() {
    setTimeout(() => {
      if (messagesContainer) {
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
      }
    }, 100);
  }

  async function sendMessage() {
    if (!inputValue.trim() || isLoading) return;

    const userMessage: V1Message = {
      id: Date.now().toString(),
      role: "user",
      content: [
        {
          text: inputValue.trim(),
        },
      ],
    };

    // Add user message to UI immediately
    messages = [...messages, userMessage];
    const currentInput = inputValue;
    inputValue = "";
    isLoading = true;
    scrollToBottom();

    try {
      const requestBody: RuntimeServiceCompleteBody = {
        useAgent: true,
        conversationId: currentConversationId || undefined,
        messages: [...messages],
      };

      const result = await $completeMutation.mutateAsync({
        instanceId,
        data: requestBody,
      });

      if (result.conversationId && !currentConversationId) {
        currentConversationId = result.conversationId;
      }

      if (result.messages) {
        messages = [...messages, ...result.messages];
        scrollToBottom();
      }

      // Refresh conversations list
      await $listConversationsQuery.refetch();
    } catch (error) {
      console.error("Failed to send message:", error);
      // Remove the optimistic user message on error
      messages = messages.filter((msg) => msg.id !== userMessage.id);
    } finally {
      isLoading = false;
    }
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      sendMessage();
    }
  }

  function startNewConversation() {
    currentConversationId = null;
    messages = [];
    inputValue = "";
  }

  function loadConversation(conversationId: string | undefined) {
    if (!conversationId) return;
    currentConversationId = conversationId;
    // Create new query for this conversation
    getConversationQuery = createRuntimeServiceGetConversation(
      instanceId,
      conversationId,
      {},
    );
  }

  function formatMessageContent(message: V1Message): string {
    if (!message.content || message.content.length === 0) return "";

    const content = message.content[0];
    if (content.text) {
      return content.text;
    }
    return "";
  }
</script>

<svelte:head>
  <title>AI Chat - Rill</title>
</svelte:head>

<div class="h-full flex bg-gray-50">
  <!-- Sidebar -->
  <div class="w-64 bg-white border-r border-gray-200 flex flex-col">
    <div class="p-4 border-b border-gray-200">
      <button
        class="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        on:click={startNewConversation}
      >
        New Chat
      </button>
    </div>

    <div class="flex-1 overflow-y-auto p-4">
      <h3 class="text-sm font-medium text-gray-500 mb-3">
        Recent Conversations
      </h3>
      {#if $listConversationsQuery.isLoading}
        <div class="text-sm text-gray-400">Loading...</div>
      {:else if conversations.length === 0}
        <div class="text-sm text-gray-400">No conversations yet</div>
      {:else}
        <div class="space-y-2">
          {#each conversations as conversation}
            <button
              class="w-full text-left p-3 rounded-lg hover:bg-gray-100 transition-colors {currentConversationId ===
              conversation.id
                ? 'bg-blue-50 border border-blue-200'
                : ''}"
              on:click={() => loadConversation(conversation.id)}
            >
              <div class="text-sm font-medium text-gray-900 truncate">
                {conversation.title || "Untitled Chat"}
              </div>
              <div class="text-xs text-gray-500 mt-1">
                {conversation.createdOn
                  ? new Date(conversation.createdOn).toLocaleDateString()
                  : ""}
              </div>
            </button>
          {/each}
        </div>
      {/if}
    </div>
  </div>

  <!-- Main Chat Area -->
  <div class="flex-1 flex flex-col">
    <!-- Header -->
    <div class="bg-white border-b border-gray-200 p-4">
      <h1 class="text-xl font-semibold text-gray-900">AI Assistant</h1>
      <p class="text-sm text-gray-500">
        Ask me anything about your data and analytics
      </p>
    </div>

    <!-- Messages -->
    <div class="flex-1 overflow-y-auto p-4" bind:this={messagesContainer}>
      {#if messages.length === 0}
        <div class="flex items-center justify-center h-full">
          <div class="text-center">
            <div class="text-gray-400 mb-4">
              <svg
                class="w-16 h-16 mx-auto"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  stroke-width="2"
                  d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
                />
              </svg>
            </div>
            <p class="text-lg font-medium text-gray-900 mb-2">
              Start a conversation
            </p>
            <p class="text-gray-500">
              Ask me to help analyze your data, create visualizations, or answer
              questions.
            </p>
          </div>
        </div>
      {:else}
        <div class="space-y-6">
          {#each messages as message}
            <div
              class="flex {message.role === 'user'
                ? 'justify-end'
                : 'justify-start'}"
            >
              <div
                class="max-w-3xl {message.role === 'user'
                  ? 'bg-blue-600 text-white'
                  : 'bg-white border border-gray-200'} rounded-lg p-4 shadow-sm"
              >
                <div class="flex items-start space-x-3">
                  <div class="flex-shrink-0">
                    {#if message.role === "user"}
                      <div
                        class="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center"
                      >
                        <svg
                          class="w-4 h-4 text-white"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                          />
                        </svg>
                      </div>
                    {:else}
                      <div
                        class="w-8 h-8 bg-gray-200 rounded-full flex items-center justify-center"
                      >
                        <svg
                          class="w-4 h-4 text-gray-600"
                          fill="none"
                          stroke="currentColor"
                          viewBox="0 0 24 24"
                        >
                          <path
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            stroke-width="2"
                            d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"
                          />
                        </svg>
                      </div>
                    {/if}
                  </div>
                  <div class="flex-1">
                    <div
                      class="text-sm font-medium {message.role === 'user'
                        ? 'text-white'
                        : 'text-gray-900'} mb-1"
                    >
                      {message.role === "user" ? "You" : "AI Assistant"}
                    </div>
                    <div
                      class="text-sm {message.role === 'user'
                        ? 'text-white'
                        : 'text-gray-800'} whitespace-pre-wrap"
                    >
                      {formatMessageContent(message)}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Input Area -->
    <div class="bg-white border-t border-gray-200 p-4">
      <div class="flex items-end space-x-4">
        <div class="flex-1">
          <textarea
            bind:value={inputValue}
            on:keydown={handleKeyDown}
            placeholder="Ask me about your data..."
            class="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
            rows="3"
            disabled={isLoading}
          />
        </div>
        <button
          on:click={sendMessage}
          disabled={!inputValue.trim() || isLoading}
          class="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          {#if isLoading}
            <svg
              class="w-4 h-4 animate-spin"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M12 6v6m0 0v6m0-6h6m-6 0H6"
              />
            </svg>
          {:else}
            Send
          {/if}
        </button>
      </div>
      <div class="text-xs text-gray-500 mt-2">
        Press Enter to send, Shift+Enter for new line
      </div>
    </div>
  </div>
</div>
