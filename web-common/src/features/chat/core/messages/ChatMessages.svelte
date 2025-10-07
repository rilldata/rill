<script lang="ts">
  import { afterUpdate } from "svelte";
  import LoadingSpinner from "../../../../components/icons/LoadingSpinner.svelte";
  import DelayedSpinner from "../../../entity-management/DelayedSpinner.svelte";
  import type { Chat } from "../chat";
  import ChatError from "./ChatError.svelte";
  import ChatMessage from "./ChatMessage.svelte";

  export let chat: Chat;
  export let layout: "sidebar" | "fullpage";

  let messagesContainer: HTMLDivElement;

  $: currentConversationStore = chat.getCurrentConversation();
  $: currentConversation = $currentConversationStore;
  $: getConversationQuery = currentConversation.getConversationQuery();

  // Loading states - access the store from the conversation instance
  $: isStreamingStore = currentConversation.isStreaming;
  $: isStreaming = $isStreamingStore;
  $: isConversationLoading = !!$getConversationQuery.isLoading;

  // Error handling
  $: streamErrorStore = currentConversation.streamError;
  $: conversationQueryError = currentConversation.getConversationQueryError();
  $: hasConversationLoadError = !!$conversationQueryError;
  $: hasStreamError = !!$streamErrorStore;

  // Data
  $: messages = $getConversationQuery.data?.conversation?.messages ?? [];

  // Auto-scroll to bottom when messages change or loading state changes
  afterUpdate(() => {
    if (messagesContainer && layout === "sidebar") {
      // For sidebar layout, scroll the messages container
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    } else if (layout === "fullpage") {
      // For fullpage layout, scroll the parent wrapper
      const parentWrapper = messagesContainer.closest(".chat-messages-wrapper");
      if (parentWrapper) {
        parentWrapper.scrollTop = parentWrapper.scrollHeight;
      }
    }
  });
</script>

<div
  class="chat-messages"
  class:sidebar={layout === "sidebar"}
  class:fullpage={layout === "fullpage"}
  bind:this={messagesContainer}
>
  {#if isConversationLoading}
    <div class="chat-loading">
      <DelayedSpinner isLoading={isConversationLoading} size="24px" />
    </div>
  {:else if hasConversationLoadError}
    <ChatError
      headline="Unable to load conversation"
      error={$conversationQueryError}
    />
  {:else if messages.length === 0}
    <div class="chat-empty">
      <!-- <div class="chat-empty-icon">💬</div> -->
      <div class="chat-empty-title">How can I help you today?</div>
      <div class="chat-empty-subtitle">Happy to help explore your data</div>
    </div>
  {:else}
    {#each messages as msg (msg.id)}
      <ChatMessage message={msg} />
    {/each}
  {/if}
  {#if isStreaming}
    <div class="response-loading">
      <LoadingSpinner size="1.2em" />
      Thinking...
    </div>
  {:else if hasStreamError}
    <ChatError
      headline="Failed to generate response"
      error={$streamErrorStore}
    />
  {/if}
</div>

<style lang="postcss">
  .chat-messages {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    background: var(--surface);
  }

  /* Sidebar layout: messages container scrolls */
  .chat-messages.sidebar {
    overflow-y: auto;
    padding: 0rem 1rem 0rem 1rem;
  }

  /* Fullpage layout: parent scrolls, content centered */
  .chat-messages.fullpage {
    padding: 1rem 1rem;
    max-width: 48rem;
    margin: 0 auto;
    width: 100%;
    min-height: 100%;
  }

  .chat-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    text-align: center;
    color: #6b7280;
  }

  /* Fullpage layout: enhanced empty state */
  .chat-messages.fullpage .chat-empty {
    padding: 4rem 2rem;
  }

  .chat-empty-title {
    font-size: 1rem;
    font-weight: 600;
    color: #374151;
    margin-bottom: 0.25rem;
  }

  .chat-messages.fullpage .chat-empty-title {
    font-size: 1.5rem;
    font-weight: 600;
    color: #111827;
    margin-bottom: 0.5rem;
  }

  .chat-empty-subtitle {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .chat-messages.fullpage .chat-empty-subtitle {
    font-size: 1rem;
    color: #6b7280;
  }

  /* Responsive: adjust fullpage empty state for mobile */
  @media (max-width: 640px) {
    .chat-messages.fullpage .chat-empty {
      padding: 2rem 1rem;
    }

    .chat-messages.fullpage .chat-empty-title {
      font-size: 1.25rem;
    }

    .chat-messages.fullpage .chat-empty-subtitle {
      font-size: 0.875rem;
    }
  }

  .response-loading {
    display: flex;
    align-items: center;
    justify-content: start;
    color: #3b82f6;
    gap: 0.5rem;
    padding: 0.5rem;
    font-size: 0.875rem;
  }
</style>
