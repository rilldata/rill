<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import UndoIcon from "@rilldata/web-common/components/icons/UndoIcon.svelte";
  import { onMount } from "svelte";
  import { base } from "../../components/editor/presets/base";
  import { debounce } from "../../lib/create-debouncer";
  import { FILE_SAVE_DEBOUNCE_TIME } from "./config";
  import { FileArtifact } from "../entity-management/file-artifacts";

  export let extensions: Extension[] = [];
  export let autoSave: boolean;
  export let disableAutoSave: boolean;
  export let fileArtifact: FileArtifact;

  let editor: EditorView;
  let container: HTMLElement;

  $: if (editor) updateEditorExtensions(extensions);
  $: ({
    hasUnsavedChanges,
    saveLocalContent,
    updateLocalContent,
    revert,
    fileQuery,
    blob,
  } = fileArtifact);

  $: ({ data } = $fileQuery);

  $: if (editor && data && data.blob !== null && data.blob !== undefined) {
    if (!editor.hasFocus && data.blob !== editor.state.doc.toString()) {
      editor.dispatch({
        changes: {
          from: 0,
          to: editor.state.doc.length,
          insert: data.blob,
          newLength: data.blob.length,
        },
        scrollIntoView: true,
      });
    }
  }

  onMount(() => {
    editor = new EditorView({
      state: EditorState.create({
        doc: blob ?? "",
        extensions: [
          // any extensions passed as props
          ...extensions,
          // establish a basic editor
          base(),
          EditorView.updateListener.of(({ docChanged, state }) => {
            if (docChanged && editor.hasFocus) {
              const latest = state.doc.toString();
              updateLocalContent(latest);
              if (!disableAutoSave && autoSave) {
                debounceSave();
              }
            }
          }),
        ],
      }),
      parent: container,
    });
  });

  function updateEditorExtensions(newExtensions: Extension[]) {
    editor.setState(
      EditorState.create({
        doc: data?.blob,
        extensions: [
          // establish a basic editor
          base(),
          // any extensions passed as props
          ...newExtensions,

          EditorView.updateListener.of(({ docChanged, state }) => {
            if (docChanged) {
              const latest = state.doc.toString();
              updateLocalContent(latest);
              if (!disableAutoSave && autoSave) {
                debounceSave();
              }
            }
          }),
        ],
      }),
    );
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      saveLocalContent().catch(console.error);
    }
  }

  $: debounceSave = debounce(saveLocalContent, FILE_SAVE_DEBOUNCE_TIME);

  function revertContent() {
    revert();
    editor.dispatch({
      changes: {
        from: 0,
        to: editor.state.doc.length,
        insert: data?.blob ?? "",
        newLength: data?.blob?.length ?? 0,
      },
      scrollIntoView: true,
    });
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<section>
  <div class="editor-container">
    <div
      bind:this={container}
      class="size-full"
      on:click={() => {
        /** give the editor focus no matter where we click */
        if (!editor.hasFocus) editor.focus();
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
        <Button disabled={!$hasUnsavedChanges} on:click={saveLocalContent}>
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
      <Switch bind:checked={autoSave} id="auto-save" small />
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
