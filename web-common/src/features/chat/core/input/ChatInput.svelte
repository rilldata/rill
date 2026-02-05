<script lang="ts">
  import { getEditorPlugins } from "@rilldata/web-common/features/chat/core/context/editor-plugins.ts";
  import { chatMounted } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { Editor } from "@tiptap/core";
  import { onMount, tick } from "svelte";
  import IconButton from "../../../../components/button/IconButton.svelte";
  import StopCircle from "../../../../components/icons/StopCircle.svelte";
  import type { ConversationManager } from "../conversation-manager";
  import type { ChatConfig } from "@rilldata/web-common/features/chat/core/types.ts";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { ArrowUp } from "lucide-svelte";

  export let conversationManager: ConversationManager;
  export let onSend: (() => void) | undefined = undefined;
  export let noMargin = false;
  export let height: string | undefined = undefined;
  export let config: ChatConfig;
  export let inline = false;

  let value = "";

  $: ({ placeholder, additionalContextStoreGetter } = config);
  $: additionalContextStore = additionalContextStoreGetter();

  $: currentConversationStore = conversationManager.getCurrentConversation();
  $: currentConversation = $currentConversationStore;
  $: getConversationQuery = currentConversation.getConversationQuery();
  $: draftMessageStore = currentConversation.draftMessage;
  $: isStreamingStore = currentConversation.isStreaming;

  $: value = $draftMessageStore;
  $: disabled = $getConversationQuery?.isLoading || $isStreamingStore;
  $: canSend = !disabled && value.trim();
  $: canCancel = $isStreamingStore;

  let element: HTMLDivElement;
  let editor: Editor;

  async function sendMessage() {
    if (!canSend) return;

    // Message handling with input focus
    try {
      await currentConversation.sendMessage($additionalContextStore, {
        onStreamStart: () => editor.commands.setContent(""),
      });
      onSend?.();
    } catch (error) {
      console.error("Failed to send message:", error);
    }

    // Let the parent component manage the input value
    await tick();
    editor.commands.focus();
  }

  function cancelStream() {
    currentConversation.cancelStream();
  }

  // Public method to focus input (can be called from parent)
  export function focusInput() {
    tick().then(() => {
      setTimeout(() => {
        editor?.commands.focus();
      }, 100);
    });
  }

  function startMention() {
    editor.commands.startMention();
  }

  function startChat(prompt: string) {
    editor.commands.setContent(prompt);
    // Wait for `value` and `canSend` to update before sending the message.`
    tick().then(sendMessage).catch(console.error);
  }

  onMount(() => {
    editor = new Editor({
      element,
      extensions: getEditorPlugins({
        placeholder,
        onSubmit: () => void sendMessage(),
      }),
      content: "",
      editorProps: {
        attributes: {
          class: config.minChatHeight,
          style: height ? `height: ${height};` : "",
        },
      },
      onTransaction: () => {
        // force re-render so `editor.isActive` works as expected
        editor = editor;
      },
      onUpdate: ({ editor }) => {
        draftMessageStore.set(editor.getText());
      },
    });

    const unsubStartChatEvent = eventBus.on("start-chat", startChat);

    chatMounted.set(true);

    return () => {
      chatMounted.set(false);
      editor.destroy();
      unsubStartChatEvent();
    };
  });
</script>

<form
  class:inline
  class="chat-input-form"
  class:no-margin={noMargin}
  on:submit|preventDefault={sendMessage}
>
  <div class="chat-input-container" bind:this={element} />
  <div class="chat-input-footer">
    <button
      class="text-base text-fg-muted"
      type="button"
      on:click={startMention}
    >
      @
    </button>
    <div class="grow"></div>
    <div>
      {#if canCancel}
        <IconButton
          ariaLabel="Cancel streaming"
          disableHover
          on:click={cancelStream}
        >
          <span class="stop-icon">
            <StopCircle size="1.2em" />
          </span>
        </IconButton>
      {:else}
        <Button
          type="primary"
          label="Send message"
          disabled={!canSend}
          square
          onClick={sendMessage}
        >
          <ArrowUp size="16px" />
        </Button>
      {/if}
    </div>
  </div>
</form>

<style lang="postcss">
  .chat-input-form {
    @apply flex flex-col gap-1 p-3 mx-4 mb-4;
    @apply border rounded-md bg-input;
    transition: border-color 0.2s;
  }

  .chat-input-form:focus-within {
    @apply border-ring-focus;
  }

  .chat-input-form.no-margin {
    margin: 0;
  }

  :global(.tiptap) {
    @apply outline-none;
    @apply text-sm leading-relaxed;
  }

  .chat-input-container {
    @apply w-full max-h-32 overflow-auto;
  }

  :global(.tiptap p.is-editor-empty:first-child::before) {
    content: attr(data-placeholder);
    @apply text-fg-muted pointer-events-none absolute;
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
</style>
