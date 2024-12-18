<script lang="ts" context="module">
</script>

<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState, Compartment } from "@codemirror/state";
  import { EditorView, type ViewUpdate } from "@codemirror/view";
  import { onMount } from "svelte";
  import { base as baseExtensions } from "../../components/editor/presets/base";
  import { FileArtifact } from "../entity-management/file-artifact";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { underlineSelection } from "./highlight-field";
  import { MergeView } from "@codemirror/merge";

  export let fileArtifact: FileArtifact;
  export let extensions: Extension[] = [];
  export let editor: EditorView | null = null;
  export let autoSave = true;

  const extensionCompartment = new Compartment();

  const {
    editorContent,
    remoteContent,
    merging,
    updateEditorContent,
    onEditorContentChange,
  } = fileArtifact;

  let parent: HTMLElement;
  let unsubscribers: Array<() => void> = [];
  let mergeView: MergeView | null = null;

  $: if (editor) updateEditorExtensions(extensions);
  $: if (fileArtifact) editor?.contentDOM.blur();

  $: if (parent) {
    if ($merging) {
      mountMergeView();
    } else {
      mountEditor();
    }
  }

  onMount(() => {
    const unsubLocal = onEditorContentChange(dispatchEditorChange);

    const unsubHighlighter = eventBus.on("highlightSelection", (refs) => {
      if (editor) underlineSelection(editor, refs);
    });

    unsubscribers.push(unsubLocal, unsubHighlighter);

    return () => {
      editor?.destroy();
      unsubscribers.forEach((unsub) => unsub());
    };
  });

  function mountMergeView() {
    editor?.destroy();

    mergeView = new MergeView({
      a: {
        doc: $editorContent ?? "",
        extensions: [
          baseExtensions(),
          ...extensions,
          EditorView.editable.of(false),
          EditorState.readOnly.of(true),
        ],
      },
      b: {
        doc: $remoteContent ?? "",
        extensions: [
          baseExtensions(),
          ...extensions,

          EditorView.editable.of(false),
          EditorState.readOnly.of(true),
        ],
      },
      parent,
    });
  }

  function mountEditor() {
    mergeView?.destroy();

    editor = new EditorView({
      state: EditorState.create({
        doc: $editorContent ?? "",
        extensions: [
          baseExtensions(),

          extensionCompartment.of([extensions]),
          EditorView.updateListener.of(listener),
        ],
      }),
      parent,
    });
  }

  function updateEditorExtensions(newExtensions: Extension[]) {
    editor?.dispatch({
      effects: extensionCompartment.reconfigure(newExtensions),
      scrollIntoView: true,
    });
  }

  function dispatchEditorChange(newContent: string) {
    const existingSelection = editor?.state.selection.ranges[0];

    editor?.dispatch({
      changes: {
        from: 0,
        to: editor.state.doc.length,
        insert: newContent,
        newLength: newContent.length,
      },
      selection: existingSelection && {
        anchor:
          existingSelection.from > newContent.length
            ? newContent.length
            : existingSelection.from,
        head:
          existingSelection.to > newContent.length
            ? newContent.length
            : existingSelection.to,
      },
    });
  }

  function listener({
    docChanged,
    state: { doc },
    view: { hasFocus },
  }: ViewUpdate) {
    if (hasFocus && docChanged) {
      updateEditorContent(doc.toString(), true, autoSave);
    }
  }
</script>

<div
  bind:this={parent}
  class="size-full overflow-y-auto"
  on:click={() => {
    /** give the editor focus no matter where we click */
    if (!editor?.hasFocus) editor?.focus();
  }}
  on:keydown={() => {
    /** no op for now */
  }}
  role="textbox"
  aria-label="Code editor"
  tabindex="0"
/>

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
