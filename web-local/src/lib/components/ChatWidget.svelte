<script lang="ts">
  import { onMount } from "svelte";
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
  import { chatOpen, chatActions } from "../stores/chat-store";

  let messages: V1Message[] = [];
  let inputValue = "";
  let isLoading = false;
  let currentConversationId: string | null = null;
  let conversations: V1Conversation[] = [];
  let messagesContainer: HTMLDivElement;
  let isLoadingConversation = false;

  const instanceId = "default";

  // API clients - initialize these at module level to avoid component initialization issues
  let completeMutation: any;
  let listConversationsQuery: any;
  let getConversationQuery: any = null;

  // Reactive statements
  $: if (listConversationsQuery && $listConversationsQuery.data?.conversations) {
    conversations = $listConversationsQuery.data.conversations;
  }

  $: if (
    getConversationQuery &&
    $getConversationQuery.data?.conversation?.messages &&
    isLoadingConversation
  ) {
    messages = $getConversationQuery.data.conversation.messages;
    isLoadingConversation = false;
    scrollToBottom();
  }

  function scrollToBottom() {
    setTimeout(() => {
      if (messagesContainer) {
        messagesContainer.scrollTop = messagesContainer.scrollHeight;
      }
    }, 100);
  }

  // Enable chat feature flag for this component
  onMount(() => {
    // Initialize API clients
    completeMutation = createRuntimeServiceComplete();
    listConversationsQuery = createRuntimeServiceListConversations(instanceId);

    // Set feature flag
    try {
      featureFlags.set(true, "chat");
    } catch (err) {
      console.warn("Could not set chat feature flag:", err);
    }

    // Restore conversation ID from localStorage if available
    const savedConversationId = localStorage.getItem('rill-chat-conversation-id');
    if (savedConversationId) {
      loadConversation(savedConversationId);
    }

    // Listen for toggle-chat event from header button
    const handleToggleChat = () => {
      try {
        chatActions.toggleChat();
      } catch (err) {
        console.error("Error toggling chat:", err);
      }
    };

    // Listen for ESC key to close chat
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape" && $chatOpen) {
        try {
          chatActions.closeChat();
        } catch (err) {
          console.error("Error closing chat:", err);
        }
      }
    };

    window.addEventListener("toggle-chat", handleToggleChat);
    window.addEventListener("keydown", handleKeyDown);

    return () => {
      window.removeEventListener("toggle-chat", handleToggleChat);
      window.removeEventListener("keydown", handleKeyDown);
    };
  });

  async function sendMessage() {
    if (!inputValue.trim() || isLoading || !completeMutation) return;

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

      console.log("Chat response:", result);

      // Set conversation ID from response for new conversations
      if (result.conversationId) {
        currentConversationId = result.conversationId;
        // Store in localStorage for persistence across page reloads
        localStorage.setItem('rill-chat-conversation-id', result.conversationId);
      }

      if (result.messages && result.messages.length > 0) {
        // Only replace messages if we got a meaningful response
        messages = result.messages;
        scrollToBottom();
      }

      // Refresh conversations list to show the new/updated conversation
      // Don't await this to prevent it from potentially interfering with the message state
      if (listConversationsQuery) {
        $listConversationsQuery.refetch();
      }
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
    isLoadingConversation = false;
    // Clear the saved conversation ID
    localStorage.removeItem('rill-chat-conversation-id');
    // Clear the current conversation query
    getConversationQuery = null;
  }

  function loadConversation(conversationId: string | undefined) {
    if (!conversationId) return;
    
    // Set the current conversation ID
    currentConversationId = conversationId;
    
    // Store in localStorage for persistence
    localStorage.setItem('rill-chat-conversation-id', conversationId);
    
    // Set loading flag to trigger message update when data arrives
    isLoadingConversation = true;
    
    // Create new query for this conversation
    try {
      getConversationQuery = createRuntimeServiceGetConversation(
        instanceId,
        conversationId,
        {},
      );
    } catch (err) {
      console.error("Error creating conversation query:", err);
      isLoadingConversation = false;
    }
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

<!-- Chat Widget Sidebar -->
{#if $chatOpen}
  <!-- Chat Widget -->
  <div
    class="fixed top-0 right-0 h-full w-96 bg-white shadow-2xl z-30 flex flex-col transform transition-transform duration-300 ease-in-out border-l border-gray-200"
  >
    <!-- Header -->
    <div
      class="bg-white border-b border-gray-200 p-4 flex items-center justify-between"
    >
      <div>
        <h2 class="text-lg font-semibold text-gray-900">AI Assistant</h2>
        <p class="text-sm text-gray-500">
          {#if currentConversationId}
            {conversations.find(c => c.id === currentConversationId)?.title || 'Current conversation'}
          {:else}
            Ask me about your data and analytics
          {/if}
        </p>
      </div>
      <button
        on:click={chatActions.closeChat}
        class="p-2 hover:bg-gray-100 rounded-md transition-colors"
        title="Close chat"
      >
        <svg
          class="w-5 h-5 text-gray-500"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M6 18L18 6M6 6l12 12"
          />
        </svg>
      </button>
    </div>

    <!-- Conversations Sidebar (Compact) -->
    <div class="border-b border-gray-200 p-3">
      <div class="flex items-center gap-2 mb-2">
        <button
          class="flex-1 px-3 py-1.5 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 transition-colors"
          on:click={startNewConversation}
        >
          New Chat
        </button>
      </div>

      {#if conversations.length > 0}
        <div class="max-h-32 overflow-y-auto">
          <div class="space-y-1">
            {#each conversations.slice(0, 3) as conversation}
              <button
                class="w-full text-left p-2 text-sm rounded-md hover:bg-gray-100 transition-colors {currentConversationId ===
                conversation.id
                  ? 'bg-blue-50 border border-blue-200'
                  : ''}"
                on:click={() => loadConversation(conversation.id)}
              >
                <div class="font-medium text-gray-900 truncate">
                  {conversation.title || "Untitled Chat"}
                </div>
                <div class="text-xs text-gray-500 flex justify-between">
                  <span>
                    {conversation.createdOn
                      ? new Date(conversation.createdOn).toLocaleDateString()
                      : ""}
                  </span>
                  {#if conversation.id === currentConversationId}
                    <span class="text-blue-600 font-medium">Active</span>
                  {/if}
                </div>
              </button>
            {/each}
          </div>
        </div>
      {/if}
    </div>

    <!-- Messages -->
    <div class="flex-1 overflow-y-auto p-4" bind:this={messagesContainer}>
      {#if messages.length === 0}
        <div class="flex items-center justify-center h-full">
          <div class="text-center">
            <div class="text-gray-400 mb-4">
              <svg
                class="w-12 h-12 mx-auto"
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
            <p class="text-sm font-medium text-gray-900 mb-1">
              Start a conversation
            </p>
            <p class="text-xs text-gray-500">
              Ask me to help analyze your data or answer questions.
            </p>
          </div>
        </div>
      {:else}
        <div class="space-y-4">
          {#each messages as message}
            <div
              class="flex {message.role === 'user'
                ? 'justify-end'
                : 'justify-start'}"
            >
              <div
                class="max-w-[85%] {message.role === 'user'
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-100 text-gray-900'} rounded-lg p-3 text-sm"
              >
                <div class="font-medium mb-1 text-xs opacity-75">
                  {message.role === "user" ? "You" : "AI Assistant"}
                </div>
                <div class="whitespace-pre-wrap">
                  {formatMessageContent(message)}
                </div>
              </div>
            </div>
          {/each}
        </div>
      {/if}
    </div>

    <!-- Input Area -->
    <div class="border-t border-gray-200 p-4">
      <div class="flex items-end gap-2">
        <div class="flex-1">
          <textarea
            bind:value={inputValue}
            on:keydown={handleKeyDown}
            placeholder="Ask me about your data..."
            class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none text-sm"
            rows="2"
            disabled={isLoading}
          />
        </div>
        <button
          on:click={sendMessage}
          disabled={!inputValue.trim() || isLoading}
          class="px-4 py-2 bg-blue-600 text-white text-sm rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
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
      <div class="text-xs text-gray-500 mt-1 flex justify-between">
        <span>Press Enter to send, Shift+Enter for new line</span>
        {#if currentConversationId}
          <span class="text-blue-600">
            Conversation: {currentConversationId.substring(0, 8)}...
          </span>
        {/if}
      </div>
    </div>
  </div>
{/if}
