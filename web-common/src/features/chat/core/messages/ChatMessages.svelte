<script lang="ts">
  import { afterUpdate } from "svelte";
  import LoadingSpinner from "../../../../components/icons/LoadingSpinner.svelte";
  import type { V1Message } from "../../../../runtime-client";
  import DelayedSpinner from "../../../entity-management/DelayedSpinner.svelte";
  import ChatMessage from "./ChatMessage.svelte";

  export let isConversationLoading = false;
  export let layout: "sidebar" | "fullpage" = "sidebar";
  export let loading = false;
  export let messages: V1Message[] = [];

  let messagesContainer: HTMLDivElement;

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
  {#if loading}
    <div class="response-loading">
      <LoadingSpinner size="1.2em" /> Thinking...
    </div>
  {/if}
</div>

<style lang="postcss">
  .chat-messages {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    background: #fafafa;
  }

  /* Sidebar layout: messages container scrolls */
  .chat-messages.sidebar {
    overflow-y: auto;
    padding: 0rem 1rem 0rem 1rem;
  }

  /* Fullpage layout: parent scrolls, content centered */
  .chat-messages.fullpage {
    padding: 0rem 1rem 0rem 1rem;
    max-width: 48rem;
    margin: 0 auto;
    width: 100%;
  }

  .chat-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    text-align: center;
    color: #6b7280;
    padding: 2rem;
  }

  .chat-empty-title {
    font-size: 1rem;
    font-weight: 600;
    color: #374151;
    margin-bottom: 0.25rem;
  }

  .chat-empty-subtitle {
    font-size: 0.75rem;
    color: #6b7280;
  }

  .response-loading {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #3b82f6;
    gap: 0.5rem;
    padding: 1rem;
    font-size: 0.875rem;
  }
</style>
