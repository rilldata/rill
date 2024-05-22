<script lang="ts">
  import type { Extension, SelectionRange } from "@codemirror/state";
  import {
    EditorState,
    Compartment,
    StateField,
    StateEffect,
  } from "@codemirror/state";
  import { Decoration, DecorationSet, EditorView } from "@codemirror/view";
  import { onMount, onDestroy } from "svelte";
  import { base as baseExtensions } from "../../components/editor/presets/base";
  import { FileArtifact } from "../entity-management/file-artifacts";
  import { get } from "svelte/store";
  import Dialog from "@rilldata/web-common/components/modal/dialog/Dialog.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import DialogFooter from "@rilldata/web-common/components/modal/dialog/DialogFooter.svelte";
  //   import {
  //   LineStatus,
  //   lineStatusesStateField,
  //   updateLineStatuses as updateLineStatusesEffect,
  // } from "./state";
  import { linter, Diagnostic } from "@codemirror/lint";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import type { Reference } from "../models/utils/get-table-references";

  const highlightField = StateField.define<DecorationSet>({
    create() {
      return Decoration.none;
    },
    update(underlines, tr) {
      underlines = underlines.map(tr.changes);
      underlines = underlines.update({
        filter: () => false,
      });

      for (let e of tr.effects)
        if (e.is(addHighlight)) {
          underlines = underlines.update({
            add: [highlightMark.range(e.value.from, e.value.to)],
          });
        }
      return underlines;
    },
    provide: (f) => EditorView.decorations.from(f),
  });

  const addHighlight = StateEffect.define<{ from: number; to: number }>({
    map: ({ from, to }, change) => ({
      from: change.mapPos(from),
      to: change.mapPos(to),
    }),
  });
  const highlightMark = Decoration.mark({ class: "cm-underline" });

  export function underlineSelection(refs: Reference[]) {
    const selections = refs.map((ref) => {
      return {
        from: ref.referenceIndex,
        to: ref.referenceIndex + ref.reference.length,
      };
    });

    const effects: StateEffect<unknown>[] = selections.map(({ from, to }) =>
      addHighlight.of({ from, to }),
    );

    if (!editor?.state.field(highlightField, false))
      effects.push(StateEffect.appendConfig.of([highlightField]));
    editor?.dispatch({ effects });
  }

  eventBus.on("highlightSelection", underlineSelection);

  export let fileArtifact: FileArtifact;
  export let extensions: Extension[] = [];
  export let autoSave: boolean = true;
  export let disableAutoSave: boolean = false;
  export let forceLocalUpdates: boolean = false;
  export let editor: EditorView | null = null;
  export let debounceSave: () => void;

  let parent: HTMLElement;
  let showWarning = false;
  let saving = false;

  $: ({
    updateLocalContent,
    localContent,
    saveLocalContent,
    remoteContent,
    revert,
    onRemoteContentChange,
  } = fileArtifact);

  const extensionCompartment = new Compartment();

  let unsubscribe: () => void;

  onMount(async () => {
    await fileArtifact.ready;

    console.log("MOUNTED");

    // Check if the file artifact has a local content
    // If it does, we want to use that as the initial content
    // Otherwise, we'll use the remote content
    editor = new EditorView({
      state: EditorState.create({
        doc: $localContent ?? $remoteContent ?? undefined,
        extensions: [
          baseExtensions(),
          extensionCompartment.of([]),
          highlightField,
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

    unsubscribe = onRemoteContentChange((newRemoteContent) => {
      if (editor && !editor.hasFocus && newRemoteContent !== null) {
        if (get(localContent) === null) {
          updateEditorContent(newRemoteContent);
        } else {
          showWarning = true;
        }
      }

      // const transaction = updateLineStatusesEffect.of({
      //   lineStatuses: lineStatuses,
      // });

      // view.dispatch({
      //   effects: [transaction],
      // });
    });

    // remoteContent.subscribe((newRemoteContent) => {
    //   if (editor && !editor.hasFocus && newRemoteContent !== null) {
    //     if (get(localContent) === null) {
    //       updateEditorContent(newRemoteContent);
    //     } else {
    //       showWarning = true;
    //     }
    //   }
    // });
  });

  onDestroy(() => {
    editor?.destroy();
    unsubscribe();
  });

  $: if (editor) updateEditorExtensions(extensions);

  // WHEN REMOTE CONTENT CHANGES
  // If there are no local changes, update the editor with the remote content
  // If there are local changes, show a warning dialog
  // $: if (editor && !editor.hasFocus && $remoteContent !== null) {
  //   if (get(localContent) === null) {
  //     updateEditorContent($remoteContent);
  //   } else {
  //     showWarning = true;
  //   }
  // }

  // WHEN LOCAL CONTENT CHANGES
  // If the editor doesn't have focus and local updates are forced
  // Update the editor with the local content
  $: if (
    forceLocalUpdates &&
    editor &&
    !editor.hasFocus &&
    $localContent !== null
  ) {
    updateEditorContent($localContent);
  }

  $: if (fileArtifact) editor?.contentDOM.blur();

  function updateEditorExtensions(newExtensions: Extension[]) {
    console.log("extensions");
    editor?.dispatch({
      effects: extensionCompartment.reconfigure(newExtensions),
      scrollIntoView: true,
    });
  }

  function updateEditorContent(newContent: string) {
    console.log("update");
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
