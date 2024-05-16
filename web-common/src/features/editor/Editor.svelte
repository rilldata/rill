<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import UndoIcon from "@rilldata/web-common/components/icons/UndoIcon.svelte";
  import type { Extension } from "@codemirror/state";
  import { EditorState, Compartment } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import { onMount, onDestroy } from "svelte";
  import { base as baseExtensions } from "../../components/editor/presets/base";
  import { debounce } from "../../lib/create-debouncer";
  import { FILE_SAVE_DEBOUNCE_TIME } from "./config";
  import { FileArtifact } from "../entity-management/file-artifacts";

  export let fileArtifact: FileArtifact;
  export let extensions: Extension[] = [];
  export let autoSave: boolean = true;
  export let disableAutoSave: boolean = false;
  export let editor: EditorView | undefined = undefined;
  export let onSave: (content: string) => void = () => {};
  export let onRevert: () => void = () => {};

  let parent: HTMLElement;

  $: ({
    hasUnsavedChanges,
    saveLocalContent,
    updateLocalContent,
    revert,
    localContent,
    remoteContent,
  } = fileArtifact);

  const extensionCompartment = new Compartment();

  onMount(async () => {
    await fileArtifact.ready;
    editor = new EditorView({
      state: EditorState.create({
        doc: $localContent ?? undefined,
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
  });

  onDestroy(() => {
    editor?.destroy();
  });

  function updateEditorExtensions(newExtensions: Extension[]) {
    editor?.dispatch({
      effects: extensionCompartment.reconfigure(newExtensions),
      scrollIntoView: true,
    });
  }

  $: if (editor) updateEditorExtensions(extensions);

  // Update the editor content when the remote content changes
  // So long as the editor doesn't have focus
  $: if (editor && $remoteContent && !editor?.hasFocus) {
    editor.dispatch({
      changes: {
        from: 0,
        to: editor.state.doc.length,
        insert: $remoteContent,
        newLength: $remoteContent?.length,
      },
      selection: editor.state.selection,
    });
  }

  $: if (fileArtifact) editor?.contentDOM.blur();

  async function handleKeydown(e: KeyboardEvent) {
    if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      await save();
    }
  }

  async function save() {
    await saveLocalContent();
    onSave($localContent);
  }

  $: debounceSave = debounce(save, FILE_SAVE_DEBOUNCE_TIME);

  function revertContent() {
    editor?.dispatch({
      changes: {
        from: 0,
        to: editor.state.doc.length,
        insert: $remoteContent,
      },
    });
    revert();
    onRevert();
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<section>
  <div class="editor-container">
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
  </div>

  <footer>
    <div class="flex gap-x-3">
      {#if !autoSave || disableAutoSave}
        <Button disabled={!$hasUnsavedChanges} on:click={debounceSave}>
          <Check size="14px" />
          Save
        </Button>

        <Button
          type="text"
          disabled={!$hasUnsavedChanges}
          on:click={revertContent}
        >
          <UndoIcon size="14px" />
          Revert changes
        </Button>
      {/if}
    </div>
    <div
      class="flex gap-x-1 items-center h-full bg-white rounded-full"
      class:hidden={disableAutoSave}
    >
      <Switch
        bind:checked={autoSave}
        id="auto-save"
        small
        on:click={() => {
          if (!autoSave) debounceSave();
        }}
      />
      <Label class="font-normal text-xs" for="auto-save">Auto-save</Label>
    </div>
  </footer>
</section>

<style lang="postcss">
  .editor-container {
    @apply size-full overflow-auto p-2 pb-0 flex flex-col;
  }

  footer {
    @apply justify-between items-center flex flex-none;
    @apply h-10 p-2 w-full rounded-b-sm border-t bg-white;
  }

  section {
    @apply size-full flex-col rounded-sm bg-white flex overflow-hidden relative;
  }
</style>
