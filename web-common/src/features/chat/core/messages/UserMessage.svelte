<!-- Renders user prompt messages. -->
<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import { extractMessageText } from "../utils";
  import { getEditorPlugins } from "@rilldata/web-common/features/chat/core/context/editor-plugins.ts";
  import { onMount } from "svelte";
  import { Editor } from "@tiptap/core";

  export let message: V1Message;

  let element: HTMLDivElement;
  let editor: Editor;

  // Message content
  $: content = extractMessageText(message);
  $: editor?.commands.setContent(content);

  // Use a readable editor instance to render the inline context component for us.
  onMount(() => {
    editor = new Editor({
      element,
      editable: false,
      extensions: getEditorPlugins({
        enableMention: true,
        placeholder: "",
        onSubmit: () => {},
      }),
      content,
      onTransaction: () => {
        // force re-render so `editor.isActive` works as expected
        editor = editor;
      },
    });

    return () => {
      editor.destroy();
    };
  });
</script>

<div class="chat-message">
  <div class="chat-message-content" bind:this={element}></div>
</div>

<style lang="postcss">
  .chat-message {
    @apply max-w-[90%] self-end;
  }

  .chat-message-content {
    @apply px-4 py-2 rounded-2xl;
    @apply text-sm leading-relaxed break-words;
    @apply bg-muted text-foreground rounded-br-lg;
  }

  :global(.chat-message-content .tiptap) {
    @apply p-0 min-h-4 outline-none;
    @apply text-sm leading-relaxed;
  }
</style>
