<script lang="ts">
  import { onMount, tick } from "svelte";
  import IconButton from "../../../../components/button/IconButton.svelte";
  import SendIcon from "../../../../components/icons/SendIcon.svelte";

  export let value = "";
  export let disabled = false;
  export let placeholder = "Ask about your data...";
  export let onSend: (message: string) => void;

  let textarea: HTMLTextAreaElement;

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  }

  async function sendMessage() {
    if (!value.trim() || disabled) return;
    const message = value;
    value = "";
    await tick();
    autoResize();
    textarea?.focus();
    onSend(message);
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
      }, 0);
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
      bind:value
      class="chat-input"
      {placeholder}
      rows="1"
      on:keydown={handleKeydown}
      on:input={autoResize}
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
