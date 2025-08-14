<script lang="ts">
  import { onMount, tick } from "svelte";
  import IconButton from "../../../../components/button/IconButton.svelte";
  import SendIcon from "../../../../components/icons/SendIcon.svelte";
  import type { Chat } from "../chat";

  export let chat: Chat;
  export let onSend: () => void;

  let textarea: HTMLTextAreaElement;
  let placeholder = "Ask about your data...";
  let newConversationDraft = "";

  $: pendingMessage = chat.pendingMessage;
  $: currentConversationStore = chat.getCurrentConversation();
  $: getConversationQuery = $currentConversationStore?.getConversationQuery();
  $: draftMessageStore = $currentConversationStore?.draftMessage;
  $: isSendingStore = $currentConversationStore?.isSending;

  $: value =
    $currentConversationStore && $draftMessageStore
      ? $draftMessageStore
      : newConversationDraft;

  $: disabled =
    $getConversationQuery?.isLoading ||
    ($currentConversationStore ? $isSendingStore : !!$pendingMessage);

  function handleInput(e: Event) {
    const target = e.target as HTMLTextAreaElement;
    const value = target.value;

    if ($currentConversationStore) {
      $currentConversationStore.draftMessage.set(value);
    } else {
      newConversationDraft = value;
    }

    autoResize();
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  }

  async function sendMessage() {
    if (!value.trim() || disabled) return;

    // Message handling with input focus
    try {
      if ($currentConversationStore) {
        // Send message to existing conversation
        await $currentConversationStore.sendMessage();
      } else {
        // No current conversation, start a new one with the input message
        if (newConversationDraft.trim()) {
          const createConversationPromise = chat.createConversation(
            newConversationDraft.trim(),
          );
          newConversationDraft = ""; // Immediately clear the draft
          await createConversationPromise;
        }
      }

      onSend();
    } catch (error) {
      console.error("Failed to send message:", error);
    }

    // Let the parent component manage the input value
    await tick();
    autoResize();
    textarea?.focus();
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
    <IconButton
      ariaLabel="Send message"
      disabled={!value.trim() || disabled}
      on:click={sendMessage}
    >
      <SendIcon
        size="1.3em"
        className={`${!value.trim() || disabled ? "text-gray-400" : "text-primary-400"}`}
      />
    </IconButton>
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
</style>
