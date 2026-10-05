<script lang="ts">
  import { EditorState } from "@codemirror/state";
  import { EditorView, type ViewUpdate } from "@codemirror/view";
  import { base as baseExtensions } from "@rilldata/web-common/components/editor/presets/base.ts";
  import { onMount } from "svelte";
  import { yaml } from "@codemirror/lang-yaml";

  let {
    contents,
    onChange,
  }: { contents: string; onChange: (newContents: string) => void } = $props();

  let parent: HTMLElement;
  let editor = $state<EditorView | null>(null);

  function mountEditor() {
    editor = new EditorView({
      state: EditorState.create({
        doc: contents ?? "",
        extensions: [
          baseExtensions(),
          yaml(),
          EditorView.updateListener.of(listener),
        ],
      }),
      parent,
    });
  }

  function listener({
    docChanged,
    state: { doc },
    view: { hasFocus },
  }: ViewUpdate) {
    if (!hasFocus || !docChanged) return;
    const newContents = doc.toString();
    if (newContents === contents) return;
    onChange(newContents);
  }

  onMount(() => {
    mountEditor();
    return () => editor?.destroy();
  });
</script>

<div
  bind:this={parent}
  class="size-full overflow-hidden"
  role="textbox"
  aria-label="codemirror editor"
  tabindex="0"
></div>

<style lang="postcss">
  :global(.cm-mergeView) {
    @apply h-full;
  }

  :global(.cm-editor) {
    padding-top: 2px;
  }

  :global(.cm-mergeViewEditor) {
    @apply overflow-y-auto;
  }
  :global(.cm-mergeViewEditors) {
    @apply h-full;
  }

  :global(.cm-mergeViewEditor:first-of-type) {
    @apply border-r;
  }
</style>
