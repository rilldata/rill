<script lang="ts">
  import type { RestoreChangesBlock } from "@rilldata/web-common/features/chat/core/messages/restore/restore-block.ts";
  import { extractMessageText } from "@rilldata/web-common/features/chat/core/utils.ts";
  import { getEditorPlugins } from "@rilldata/web-common/features/chat/core/context/editor-plugins.ts";
  import { onMount } from "svelte";
  import { Editor } from "@tiptap/core";

  export let block: RestoreChangesBlock;

  let element: HTMLDivElement;
  let editor: Editor;

  // Message content
  $: restoredPrompt = extractMessageText(block.restoredMessage);
  $: content = `Restored: "${restoredPrompt}"`;
  $: editor?.commands.setContent(content);

  // Use a readable editor instance to render the inline context component for us.
  onMount(() => {
    editor = new Editor({
      element,
      editable: false,
      extensions: getEditorPlugins({
        placeholder: "",
        onSubmit: () => {},
      }),
      content,
    });

    return () => {
      editor.destroy();
    };
  });
</script>

<div bind:this={element}></div>
