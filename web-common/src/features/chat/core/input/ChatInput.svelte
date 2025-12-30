<script lang="ts">
  import { getEditorPlugins } from "@rilldata/web-common/features/chat/core/context/inline-context-plugins.ts";
  import { chatMounted } from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store.ts";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
  import { Editor } from "@tiptap/core";
  import { onMount, tick } from "svelte";
  import IconButton from "../../../../components/button/IconButton.svelte";
  import SendIcon from "../../../../components/icons/SendIcon.svelte";
  import StopCircle from "../../../../components/icons/StopCircle.svelte";
  import type { ConversationManager } from "../conversation-manager";
  import type { ChatConfig } from "@rilldata/web-common/features/chat/core/types.ts";

  export let conversationManager: ConversationManager;
  export let onSend: (() => void) | undefined = undefined;
  export let noMargin = false;
  export let height: string | undefined = undefined;
  export let config: ChatConfig;

  let value = "";

  $: ({ placeholder, additionalContextStoreGetter, enableMention } = config);
  $: additionalContextStore = additionalContextStoreGetter();

  $: currentConversationStore = conversationManager.getCurrentConversation();
  $: currentConversation = $currentConversationStore;
  $: getConversationQuery = currentConversation.getConversationQuery();
  $: draftMessageStore = currentConversation.draftMessage;
  $: isStreamingStore = currentConversation.isStreaming;

  let streamStartUnsub: (() => void) | undefined = undefined;
  $: {
    streamStartUnsub?.();
    streamStartUnsub = currentConversation.on("stream-start", () => {
      editor.commands.setContent("");
    });
  }

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
      await currentConversation.sendMessage($additionalContextStore);
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
        enableMention,
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
  class="chat-input-form"
  class:no-margin={noMargin}
  on:submit|preventDefault={sendMessage}
>
  <div class="chat-input-container" bind:this={element} />
  <div class="chat-input-footer">
    {#if enableMention}
      <button class="text-base ml-1" type="button" on:click={startMention}>
        @
      </button>
    {/if}
    <div class="grow"></div>
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
    @apply flex flex-col gap-1 py-1 mx-4 mb-4;
    @apply bg-background border rounded-md;
    transition: border-color 0.2s;
  }

  .chat-input-form:focus-within {
    @apply border-primary-400;
  }

  .chat-input-form.no-margin {
    margin: 0;
  }

  :global(.tiptap) {
    @apply px-2 py-2 outline-none;
    @apply text-sm leading-relaxed;
  }

  .chat-input-container {
    @apply w-full max-h-32 overflow-auto;
  }

  :global(.tiptap p.is-editor-empty:first-child::before) {
    content: attr(data-placeholder);
    @apply text-gray-400 pointer-events-none absolute;
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
    @apply flex flex-row px-1;
  }
</style>
