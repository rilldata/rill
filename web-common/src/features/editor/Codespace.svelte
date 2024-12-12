<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState, Compartment } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
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
  } = fileArtifact;

  onMount(() => {
    if ($merging) {
      mountMergeView();
    } else {
      mountEditor();
    }

    const unsubscribeRemoteContent = onRemoteContentChange(
      (newRemoteContent) => {
        if (editor && !editor.hasFocus) {
          if (editor.state.doc.toString() === newRemoteContent) return;

          if (autoSave) {
            updateEditorContent(newRemoteContent);
          }
        }
      },
    );

    const unsubscribeLocalContent = onLocalContentChange((content) => {
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

    const unsubscribeHighlighter = eventBus.on("highlightSelection", (refs) => {
      if (editor) underlineSelection(editor, refs);
    });

    unsubscribers.push(
      unsubscribeRemoteContent,
      unsubscribeLocalContent,
      unsubscribeHighlighter,
    );

    return () => {
      editor?.destroy();
      unsubscribers.forEach((unsub) => unsub());
    };
  });

  // $: if (editor) updateEditorExtensions(extensions);
  $: if (fileArtifact) editor?.contentDOM.blur();

  // function updateEditorExtensions(newExtensions: Extension[]) {
  //   editor?.dispatch({
  //     effects: extensionCompartment.reconfigure(newExtensions),
  //     scrollIntoView: true,
  //   });
  // }

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

    const mergeExtensions = [
      baseExtensions(),
      ...extensions,
      EditorView.editable.of(false),
      EditorState.readOnly.of(true),
    ];

    merging.set(true);
    editor?.destroy();

    mergeView = new MergeView({
      a: {
        doc: $localContent,
        extensions: mergeExtensions,
      },
      b: {
        doc: $remoteContent,
        extensions: mergeExtensions,
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
          extensionCompartment.of([extensions]),
          EditorView.updateListener.of(
            ({ docChanged, state: { doc }, view: { hasFocus } }) => {
              editorHasFocus = hasFocus;

              if (hasFocus && docChanged) {
                updateLocalContent(doc.toString());
              }
            },
          ),
        ],
      }),
      parent,
    });
  }
</script>

<div
  bind:this={parent}
  class="size-full"
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
