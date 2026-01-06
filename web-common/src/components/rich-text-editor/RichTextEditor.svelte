<script lang="ts">
  import { Editor } from "@tiptap/core";
  import { onMount } from "svelte";
  import StarterKit from "@tiptap/starter-kit";
  import { getEditorExtensions } from "./extensions";

  export let content: string = "";
  export let placeholder: string = "";
  export let onUpdate: ((content: string) => void) | undefined = undefined;
  export let metricsViewName: string | undefined = undefined;
  export let availableMeasures: string[] = [];

  let element: HTMLDivElement;
  let editor: Editor | null = null;
  let isUpdating = false;
  let hasRecreatedForMeasures = false;

  $: if (editor && content !== editor.getHTML() && !isUpdating) {
    isUpdating = true;
    editor.commands.setContent(content, false);
    isUpdating = false;
  }

  // Recreate editor once when measures become available (if editor was created without them)
  $: if (
    editor &&
    metricsViewName &&
    availableMeasures &&
    availableMeasures.length > 0 &&
    !hasRecreatedForMeasures
  ) {
    const currentContent = editor.getHTML();
    const currentSelection = editor.state.selection;
    
    editor.destroy();
    editor = null;
    hasRecreatedForMeasures = true;

    // Recreate editor with measures
    setTimeout(() => {
      if (!element) return;

      try {
        editor = new Editor({
          element,
          extensions: getEditorExtensions({
            placeholder,
            metricsViewName,
            availableMeasures,
          }),
          content: currentContent,
          editorProps: {
            attributes: {
              class: "rich-text-editor",
            },
          },
          onUpdate: ({ editor }) => {
            if (isUpdating) return;
            const html = editor.getHTML();
            onUpdate?.(html);
          },
        });

        // Restore selection if possible
        try {
          const docSize = editor.state.doc.content.size;
          if (currentSelection.from <= docSize && currentSelection.to <= docSize) {
            editor.commands.setTextSelection({
              from: currentSelection.from,
              to: currentSelection.to,
            });
          }
        } catch {
          // Selection restore failed, ignore
        }
      } catch (error) {
        console.error("RichTextEditor: Failed to recreate editor with measures", error);
      }
    }, 0);
  }

  onMount(() => {
    if (!element) {
      console.error("RichTextEditor: element not found");
      return;
    }

    try {
      const extensions = getEditorExtensions({
        placeholder,
        metricsViewName,
        availableMeasures,
      });

      editor = new Editor({
        element,
        extensions,
        content: content || "",
        editorProps: {
          attributes: {
            class: "rich-text-editor",
          },
        },
        onUpdate: ({ editor }) => {
          if (isUpdating) return;
          const html = editor.getHTML();
          onUpdate?.(html);
        },
      });

      // Mark if we already have measures
      if (availableMeasures && availableMeasures.length > 0) {
        hasRecreatedForMeasures = true;
      }
    } catch (error) {
      console.error("RichTextEditor: Failed to initialize editor", error);
    }

    return () => {
      editor?.destroy();
      editor = null;
      hasRecreatedForMeasures = false;
    };
  });

  export function getContent(): string {
    return editor?.getHTML() ?? "";
  }

  export function getText(): string {
    return editor?.getText() ?? "";
  }

  export function focus(): void {
    editor?.commands.focus();
  }
</script>

<div bind:this={element} class="rich-text-editor-wrapper" />

<style lang="postcss">
  :global(.rich-text-editor) {
    @apply px-2 py-2 outline-none;
    @apply text-sm leading-relaxed;
    @apply min-h-[200px];
  }

  :global(.rich-text-editor p.is-editor-empty:first-child::before) {
    content: attr(data-placeholder);
    @apply text-gray-400 pointer-events-none float-left h-0;
  }

  :global(.rich-text-editor h1) {
    @apply text-2xl font-bold mb-2 mt-4;
  }

  :global(.rich-text-editor h1:first-child) {
    @apply mt-0;
  }

  :global(.rich-text-editor h2) {
    @apply text-xl font-bold mb-2 mt-3;
  }

  :global(.rich-text-editor h2:first-child) {
    @apply mt-0;
  }

  :global(.rich-text-editor h3) {
    @apply text-lg font-semibold mb-1 mt-2;
  }

  :global(.rich-text-editor h3:first-child) {
    @apply mt-0;
  }

  :global(.rich-text-editor p) {
    @apply mb-2;
  }

  :global(.rich-text-editor p:last-child) {
    @apply mb-0;
  }

  :global(.rich-text-editor ul),
  :global(.rich-text-editor ol) {
    @apply pl-6 mb-2;
  }

  :global(.rich-text-editor ul) {
    @apply list-disc;
  }

  :global(.rich-text-editor ol) {
    @apply list-decimal;
  }

  :global(.rich-text-editor li) {
    @apply mb-1;
  }

  :global(.rich-text-editor strong) {
    @apply font-bold;
  }

  :global(.rich-text-editor em) {
    @apply italic;
  }

  .rich-text-editor-wrapper {
    @apply w-full border border-gray-300 rounded-sm;
    @apply bg-white;
  }

  .rich-text-editor-wrapper:focus-within {
    @apply border-primary-400 ring-1 ring-primary-400;
  }
</style>

