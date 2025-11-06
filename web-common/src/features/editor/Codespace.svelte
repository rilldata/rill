<script lang="ts">
  import { beforeNavigate } from "$app/navigation";
  import { MergeView } from "@codemirror/merge";
  import type { Extension } from "@codemirror/state";
  import { Compartment, EditorSelection, EditorState } from "@codemirror/state";
  import { EditorView, type ViewUpdate } from "@codemirror/view";
  import { onMount } from "svelte";
  import { base as baseExtensions } from "../../components/editor/presets/base";
  import { FileArtifact } from "../entity-management/file-artifact";

  export let fileArtifact: FileArtifact;
  export let extensions: Extension[] = [];
  export let editor: EditorView | null = null;
  export let autoSave = true;

  const extensionCompartment = new Compartment();

  const {
    editorContent,
    remoteContent,
    merging,
    snapshot,
    updateEditorContent,
    onEditorContentChange,
    saveSnapshot,
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

    unsubscribers.push(unsubLocal);

    return () => {
      editor?.destroy();
      unsubscribers.forEach((unsub) => unsub());
    };
  });

  beforeNavigate(() => {
    if (!editor) return;
    saveSnapshot(editor);
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

    const { selection, scroll } = $snapshot;

    editor = new EditorView({
      state: EditorState.create({
        doc: $editorContent ?? "",
        extensions: [
          baseExtensions(),
          extensionCompartment.of([extensions]),
          EditorView.updateListener.of(listener),
        ],
        selection: EditorSelection.create(
          [
            EditorSelection.range(
              Math.min(
                $editorContent?.length ?? Infinity,
                selection?.ranges[0].anchor ?? 0,
              ),
              Math.min(
                $editorContent?.length ?? Infinity,
                selection?.ranges[0].head ?? 0,
              ),
            ),
          ],
          0,
        ),
      }),
      parent,
      scrollTo: scroll,
    });

    if (selection) editor?.focus();
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
  class="size-full overflow-hidden"
  role="textbox"
  aria-label="codemirror editor"
  tabindex="0"
/>

<style lang="postcss">
  @reference "tailwindcss";

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
