<script lang="ts">
  import { getExploreContext } from "@rilldata/web-common/features/chat/core/context/explore-context.ts";
  import ChatInputTextarea from "@rilldata/web-common/features/chat/core/input/ChatInputTextarea.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { onMount, tick } from "svelte";
  import IconButton from "../../../../components/button/IconButton.svelte";
  import SendIcon from "../../../../components/icons/SendIcon.svelte";
  import StopCircle from "../../../../components/icons/StopCircle.svelte";
  import type { ConversationManager } from "../conversation-manager";

  export let conversationManager: ConversationManager;
  export let onSend: (() => void) | undefined = undefined;
  export let noMargin = false;
  export let height: string | undefined = undefined;

  let textarea: ChatInputTextarea;
  let placeholder = "Ask about your data...";

  $: currentConversationStore = conversationManager.getCurrentConversation();
  $: currentConversation = $currentConversationStore;
  $: getConversationQuery = currentConversation.getConversationQuery();
  $: draftMessageStore = currentConversation.draftMessage;
  $: isStreamingStore = currentConversation.isStreaming;

  $: value = $draftMessageStore;
  $: disabled = $getConversationQuery?.isLoading || $isStreamingStore;
  $: canSend = !disabled && value.trim();
  $: canCancel = $isStreamingStore;

  const context = getExploreContext();

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
      textarea.setPrompt("");
      await currentConversation.sendMessage($context);
      onSend?.();
    } catch (error) {
      console.error("Failed to send message:", error);
    }

    // Let the parent component manage the input value
    await tick();
    textarea.focusEditor();
  }

  function cancelStream() {
    currentConversation.cancelStream();
  }

  // Public method to focus input (can be called from parent)
  export function focusInput() {
    tick().then(() => {
      setTimeout(() => {
        textarea.focusEditor();
      }, 100);
    });
  }

  onMount(() =>
    eventBus.on("start-chat", (prompt) => textarea.setPrompt(prompt)),
  );
</script>

<form
  class="chat-input-form"
  class:no-margin={noMargin}
  on:submit|preventDefault={sendMessage}
>
  <div class="w-full">
    <ChatInputTextarea
      bind:this={textarea}
      onChange={(newValue) => draftMessageStore.set(newValue)}
    />
  </div>
  <div class="chat-input-footer">
    <div class="chat-input-dashboard-scope"></div>
    <div>
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
          <SendIcon size="1.3em" disabled={!canSend} />
        </IconButton>
      {/if}
    </div>
  </div>
</form>

<style lang="postcss">
  .chat-input-form {
    @apply flex flex-col gap-1 p-1 mx-4;
    @apply bg-background border rounded-md;
    transition: border-color 0.2s;
  }

  .chat-input-form:focus-within {
    @apply border-primary-400;
  }

  .chat-input-form.no-margin {
    margin: 0;
  }

  .chat-input {
    @apply w-full;
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

  .chat-input.fixed-height {
    min-height: unset;
    max-height: unset;
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

  .chat-input-footer {
    @apply flex flex-row;
  }

  .chat-input-dashboard-scope {
    @apply flex flex-row items-center w-full mx-1 gap-1;
    @apply text-xs text-muted-foreground;
  }
</style>
