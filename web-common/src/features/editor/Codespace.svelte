<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState, Compartment } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import { onMount, onDestroy } from "svelte";
  import { base as baseExtensions } from "../../components/editor/presets/base";
  import { FileArtifact } from "../entity-management/file-artifacts";
  import { get } from "svelte/store";
  import Dialog from "@rilldata/web-common/components/modal/dialog/Dialog.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import DialogFooter from "@rilldata/web-common/components/modal/dialog/DialogFooter.svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { underlineSelection } from "./highlight-field";

  export let fileArtifact: FileArtifact;
  export let extensions: Extension[] = [];
  export let autoSave: boolean = true;
  export let disableAutoSave: boolean = false;
  export let forceLocalUpdates: boolean = false;
  export let editor: EditorView | null = null;
  export let debounceSave: () => void;

  const extensionCompartment = new Compartment();

  let parent: HTMLElement;
  let showWarning = false;
  let saving = false;

  let unsubscribers: Array<() => void> = [];

  $: ({
    updateLocalContent,
    localContent,
    saveLocalContent,
    remoteContent,
    revert,
    onRemoteContentChange,
    onLocalContentChange,
  } = fileArtifact);

  onMount(async () => {
    await fileArtifact.ready;

    // Check if the file artifact has a local content
    // If it does, we want to use that as the initial content
    // Otherwise, we'll use the remote content
    editor = new EditorView({
      state: EditorState.create({
        doc: $localContent ?? $remoteContent ?? undefined,
        extensions: [
          baseExtensions(),
          extensionCompartment.of([]),
          EditorView.updateListener.of(({ docChanged, state: { doc } }) => {
            if (editor?.hasFocus && docChanged) {
              updateLocalContent(doc.toString());

              if (!disableAutoSave && autoSave) debounceSave();
            }
          }),
        ],
      }),
      parent,
    });

    const unsubscribeRemoteContent = onRemoteContentChange(
      (newRemoteContent) => {
        if (editor && !editor.hasFocus && newRemoteContent !== null) {
          if (!get(localContent)) {
            updateEditorContent(newRemoteContent);
          } else {
            showWarning = true;
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
      )
        updateEditorContent(content);
    });

    const unsubscribeHighlighter = eventBus.on("highlightSelection", (refs) => {
      if (editor) underlineSelection(editor, refs);
    });

    unsubscribers.push(
      unsubscribeRemoteContent,
      unsubscribeLocalContent,
      unsubscribeHighlighter,
    );
  });

  onDestroy(() => {
    editor?.destroy();
    unsubscribers.forEach((unsub) => unsub());
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
    editor?.dispatch({
      changes: {
        from: 0,
        to: editor.state.doc.length,
        insert: newContent,
        newLength: newContent.length,
      },
      selection: editor.state.selection,
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
  tabindex="0"
/>

{#if showWarning}
  <Dialog
    on:cancel={() => (showWarning = false)}
    size="sm"
    useContentForMinSize
    focusTriggerOnClose={false}
    showCancel={false}
  >
    <svelte:fragment slot="title">
      File update received remotely
    </svelte:fragment>
    <DialogFooter slot="footer">
      <Button
        type="secondary"
        loading={saving}
        on:click={async () => {
          saving = true;
          await saveLocalContent();
          saving = false;
          showWarning = false;
        }}
      >
        Save local changes
      </Button>
      <Button
        type="primary"
        on:click={() => {
          showWarning = false;
          revert();
          if ($remoteContent !== null) updateEditorContent($remoteContent);
        }}
      >
        Accept remote changes
      </Button>
    </DialogFooter>
  </Dialog>
{/if}
