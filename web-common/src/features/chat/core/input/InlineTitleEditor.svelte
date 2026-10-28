<!-- Single line chat editor used to edit a conversation title with inline chat references. -->
<script lang="ts">
  import {
    getEditorPlugins,
    SharedEditorStore,
  } from "@rilldata/web-common/features/chat/core/context/editor-plugins.svelte.ts";
  import { Editor } from "@tiptap/core";
  import { onMount } from "svelte";

  let {
    value,
    onSave,
    onCancel,
  }: {
    // Raw title, with chat references as markup.
    value: string;
    onSave: (title: string) => void;
    onCancel: () => void;
  } = $props();

  let container: HTMLDivElement;
  let editor: Editor;

  // Tracks whether a picker or a reference dropdown is open,
  // in which case Escape and outside clicks belong to them rather than the title edit.
  const sharedEditorStore = new SharedEditorStore();
  // Created during component initialization since the picker captures the parent contexts.
  const extensions = getEditorPlugins({
    placeholder: "",
    onSubmit: save,
    sharedEditorStore,
  });

  onMount(() => {
    editor = new Editor({
      element: container,
      extensions,
      content: value,
      autofocus: "end",
    });
    return () => editor.destroy();
  });

  function save() {
    // Titles are single line, so join paragraphs added with Shift-Enter.
    onSave(editor.getText({ blockSeparator: " " }).trim());
  }

  function handleKeyDown(e: KeyboardEvent) {
    // The suggestion plugin handles Escape to close the picker, which prevents the default.
    if (e.key !== "Escape" || e.defaultPrevented) return;
    if (sharedEditorStore.contextOpen) return;
    onCancel();
  }

  // Uses pointerdown during capture so that the picker is still marked as open
  // when one of its options is clicked; it closes before the click reaches the document.
  function handleDocumentPointerDown(e: PointerEvent) {
    if (!(e.target instanceof Node) || container.contains(e.target)) return;
    if (sharedEditorStore.contextOpen) return;
    save();
  }
</script>

<svelte:document onpointerdowncapture={handleDocumentPointerDown} />

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="title-editor" bind:this={container} onkeydown={handleKeyDown}></div>

<style lang="postcss">
  .title-editor {
    @apply w-full min-w-0 border-b border-primary-500;
  }

  .title-editor :global(.tiptap) {
    @apply text-xs leading-normal text-fg-primary;
  }
</style>
