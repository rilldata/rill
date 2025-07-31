<script lang="ts">
  import { afterUpdate } from "svelte";
  import AlertCircle from "../../../components/icons/AlertCircle.svelte";
  import LoadingSpinner from "../../../components/icons/LoadingSpinner.svelte";
  import DelayedSpinner from "../../entity-management/DelayedSpinner.svelte";
  import { error, loading, messages } from "../chat-store";
  import ChatMessage from "./ChatMessage.svelte";

  export let isConversationLoading = false;

  let messagesContainer: HTMLDivElement;

  // Auto-scroll to bottom when messages change or loading state changes
  afterUpdate(() => {
    if (messagesContainer) {
      messagesContainer.scrollTop = messagesContainer.scrollHeight;
    }
  });
</script>

<div class="chat-messages" bind:this={messagesContainer}>
  {#if isConversationLoading}
    <div class="chat-loading">
      <DelayedSpinner isLoading={isConversationLoading} size="24px" />
    </div>
  {:else if $messages.length === 0}
    <div class="chat-empty">
      <!-- <div class="chat-empty-icon">💬</div> -->
      <div class="chat-empty-title">How can I help you today?</div>
      <div class="chat-empty-subtitle">Happy to help explore your data</div>
    </div>
  {:else}
    {#each $messages as msg (msg.id)}
      <ChatMessage message={msg} />
    {/each}
  {/if}
  {#if $loading}
    <div class="response-loading">
      <LoadingSpinner size="1.2em" /> Thinking...
    </div>
  {/if}
  {#if $error}
    <div class="chat-error">
      <AlertCircle size="1.2em" />
      {$error}
    </div>
  {/if}
</div>

<style lang="postcss">
  .chat-messages {
    flex: 1;
    overflow-y: auto;
    padding: 0rem 1rem 0rem 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    background: #fafafa;
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

  .chat-error {
    display: flex;
    align-items: center;
    justify-content: center;
    color: #dc2626;
    gap: 0.5rem;
    padding: 1rem;
    font-size: 0.875rem;
    background: #fef2f2;
    border: 1px solid #fecaca;
    border-radius: 0.5rem;
    margin: 0.5rem;
  }
</style>
