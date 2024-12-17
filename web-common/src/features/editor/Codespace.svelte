<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState, Compartment } from "@codemirror/state";
  import { EditorView, ViewUpdate } from "@codemirror/view";
  import { onMount } from "svelte";
  import { base as baseExtensions } from "../../components/editor/presets/base";
  import { FileArtifact } from "../entity-management/file-artifact";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { underlineSelection } from "./highlight-field";
  import { MergeView } from "@codemirror/merge";

  export let fileArtifact: FileArtifact;
  export let extensions: Extension[] = [];
  export let autoSave: boolean = true;
  export let forceLocalUpdates: boolean = false;
  export let editor: EditorView | null = null;
  export let editorHasFocus = false;

  const extensionCompartment = new Compartment();
  const editable = new Compartment();

  let parent: HTMLElement;
  let unsubscribers: Array<() => void> = [];
  let mergeView: MergeView | null = null;

  const {
    remoteContent,
    localContent,
    merging,
    updateLocalContent,
    onRemoteContentChange,
    onLocalContentChange,
    saveState: { saving },
  } = fileArtifact;

  onMount(() => {
    if ($merging) {
      mountMergeView();
    } else {
      mountEditor();
    }

    const unsubRemote = onRemoteContentChange((newRemoteContent) => {
      if (editor && !editor.hasFocus) {
        if (editor.state.doc.toString() === newRemoteContent) return;

        if (autoSave || $localContent === null) {
          updateEditorContent(newRemoteContent);
        }
      }
    });

    const unsubLocal = onLocalContentChange((content) => {
      if (content === null && $remoteContent !== null) {
        updateEditorContent($remoteContent);
      } else if (
        forceLocalUpdates &&
        content !== null &&
        editor &&
        !editor.hasFocus
      ) {
        updateEditorContent(content);
      }
    });

    const unsubHighlighter = eventBus.on("highlightSelection", (refs) => {
      if (editor) underlineSelection(editor, refs);
    });

    unsubscribers.push(unsubRemote, unsubLocal, unsubHighlighter);

    return () => {
      editor?.destroy();
      unsubscribers.forEach((unsub) => unsub());
    };
  });

  $: if (editor) updateEditorExtensions(extensions);
  $: if (fileArtifact) editor?.contentDOM.blur();

  function updateEditorExtensions(newExtensions: Extension[]) {
    editor?.dispatch({
      effects: extensionCompartment.reconfigure(newExtensions),
      scrollIntoView: true,
    });
  }

  function updateEditorContent(newContent: string) {
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

  export function mountMergeView() {
    if (!$localContent || !$remoteContent) return;

    merging.set(true);
    editor?.destroy();

    mergeView = new MergeView({
      a: {
        doc: $localContent,
        extensions: [
          baseExtensions(),
          ...extensions,
          editable.of([
            EditorView.editable.of(true),
            EditorState.readOnly.of(false),
          ]),
          EditorView.updateListener.of(listener),
        ],
      },
      b: {
        doc: $remoteContent,
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

  export function mountEditor() {
    mergeView?.destroy();

    const initialContent = $localContent ?? $remoteContent ?? "";

    editor = new EditorView({
      state: EditorState.create({
        doc: initialContent,
        extensions: [
          baseExtensions(),
          editable.of([
            EditorView.editable.of(true),
            EditorState.readOnly.of(false),
          ]),
          extensionCompartment.of([extensions]),
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
    editorHasFocus = hasFocus;

    if (hasFocus && docChanged) {
      if (!autoSave && $saving) return;
      updateLocalContent(doc.toString());
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
  :global(.cm-mergeViewEditors) {
    @apply h-full;
  }

  :global(.cm-mergeViewEditor:first-of-type) {
    @apply border-r;
  }
</style>
