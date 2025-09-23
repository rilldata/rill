<script lang="ts">
  import { onMount, tick } from "svelte";
  import IconButton from "../../../../components/button/IconButton.svelte";
  import SendIcon from "../../../../components/icons/SendIcon.svelte";
  import StopCircle from "../../../../components/icons/StopCircle.svelte";
  import type { Chat } from "../chat";

  export let chat: Chat;
  export let onSend: () => void;

  let textarea: HTMLTextAreaElement;
  let placeholder = "Ask about your data...";

  $: currentConversationStore = chat.getCurrentConversation();
  $: currentConversation = $currentConversationStore;
  $: getConversationQuery = currentConversation.getConversationQuery();
  $: draftMessageStore = currentConversation.draftMessage;
  $: isStreamingStore = currentConversation.isStreaming;

  $: value = $draftMessageStore;
  $: disabled = $getConversationQuery?.isLoading || $isStreamingStore;
  $: canSend = !disabled && value.trim();
  $: canCancel = $isStreamingStore;

  function handleInput(e: Event) {
    const target = e.target as HTMLTextAreaElement;
    const value = target.value;
    draftMessageStore.set(value);
    autoResize();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      if (!$isStreamingStore) {
        sendMessage();
      }
    }
  }

  async function sendMessage() {
    if (!canSend) return;

    // Message handling with input focus
    try {
      await currentConversation.sendMessage();
      onSend();
    } catch (error) {
      console.error("Failed to send message:", error);
    }

    // Let the parent component manage the input value
    await tick();
    autoResize();
    textarea?.focus();
  }

  function cancelStream() {
    currentConversation.cancelStream();
  }

  function autoResize() {
    if (textarea) {
      textarea.style.height = "auto";
      textarea.style.height = textarea.scrollHeight + "px";
    }
  }

  // Public method to focus input (can be called from parent)
  export function focusInput() {
    tick().then(() => {
      setTimeout(() => {
        textarea?.focus();
      }, 100);
    });
  }

  onMount(() => {
    autoResize();
  });

  // Auto-resize when value changes
  $: if (textarea && value !== undefined) {
    autoResize();
  }
</script>

<form class="chat-input-form" on:submit|preventDefault={sendMessage}>
  <div class="chat-input-container">
    <textarea
      bind:this={textarea}
      {value}
      class="chat-input"
      {placeholder}
      rows="1"
      on:keydown={handleKeydown}
      on:input={handleInput}
    />
    {#if canCancel}
      <IconButton ariaLabel="Cancel streaming" on:click={cancelStream}>
        <span class="stop-icon">
          <StopCircle size="1.2em" />
        </span>
      </IconButton>
    {:else}
      <IconButton
        ariaLabel="Send message"
        disabled={!canSend}
        on:click={sendMessage}
      >
        <SendIcon
          size="1.3em"
          className={canSend ? "text-primary-400" : "text-gray-400"}
        />
      </IconButton>
    {/if}
  </div>
</form>

<style lang="postcss">
  .chat-input-form {
    padding: 1rem 1rem 0rem 1rem;
    background: #fafafa;
  }

  .chat-input-container {
    display: flex;
    align-items: flex-end;
    gap: 0.25rem;
    background: #ffffff;
    border: 1px solid #d1d5db;
    border-radius: 0.75rem;
    padding: 0.25rem;
    transition: border-color 0.2s;
  }

  .chat-input-container:focus-within {
    @apply border-primary-400;
  }

  .chat-input {
    flex: 1;
    border: none;
    background: transparent;
    font-size: 0.875rem;
    line-height: 1.4;
    outline: none;
    resize: none;
    min-height: 1.75rem;
    max-height: 6rem;
    padding: 0.25rem;
    font-family: inherit;
    overflow-y: auto;
  }

  .chat-input::placeholder {
    color: #9ca3af;
  }

  .chat-input:disabled {
    color: #9ca3af;
    cursor: not-allowed;
  }

  .stop-icon {
    color: #9ca3af; /* gray-400 base */
    display: flex;
    align-items: center;
    justify-content: center;
    transition:
      transform 120ms ease,
      color 160ms ease,
      filter 160ms ease;
    will-change: transform;
  }
  .stop-icon:hover {
    color: #6b7280; /* gray-500 on hover */
    transform: scale(1.04);
    filter: drop-shadow(0 1px 0 rgba(0, 0, 0, 0.02));
  }
  .stop-icon:active {
    transform: scale(0.97);
  }
</style>
