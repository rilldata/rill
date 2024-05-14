<script lang="ts">
  import type { Extension } from "@codemirror/state";
  import { EditorState } from "@codemirror/state";
  import { EditorView } from "@codemirror/view";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import UndoIcon from "@rilldata/web-common/components/icons/UndoIcon.svelte";
  import { createEventDispatcher, onMount } from "svelte";
  import { bindEditorEventsToDispatcher } from "../../components/editor/dispatch-events";
  import { base } from "../../components/editor/presets/base";
  import { debounce } from "../../lib/create-debouncer";
  import { FILE_SAVE_DEBOUNCE_TIME } from "./config";

  const dispatch = createEventDispatcher();

  export let remoteContent: string;
  export let localContent: string | null;
  export let extensions: Extension[] = [];
  export let autoSave: boolean;
  export let disableAutoSave: boolean;
  export let hasUnsavedChanges: boolean;
  export let editor: EditorView | undefined = undefined;
  export let key: string;

  let container: HTMLElement;

  $: updateEditorContents(remoteContent);
  $: if (editor) updateEditorExtensions(extensions);

  onMount(() => {
    editor = new EditorView({
      state: EditorState.create({
        doc: remoteContent,
        extensions: [
          // any extensions passed as props
          ...extensions,
          // establish a basic editor
          base(),
          // this will catch certain events and dispatch them to the parent
          bindEditorEventsToDispatcher(dispatch),
        ],
      }),
      parent: container,
    });
  });

  function updateEditorExtensions(newExtensions: Extension[]) {
    if (!editor) return;
    editor.setState(
      EditorState.create({
        doc: remoteContent,
        extensions: [
          // establish a basic editor
          base(),
          // any extensions passed as props
          ...newExtensions,
          EditorView.updateListener.of((v) => {
            if (v.focusChanged && v.view.hasFocus) {
              dispatch("receive-focus");
            }
            if (v.docChanged) {
              localContent = v.state.doc.toString();

              if (!disableAutoSave && autoSave) debounceSave();
            }
          }),
        ],
      }),
    );
  }

  $: if (key) {
    // When the key changes, unfocus the Editor so that the update is dispatched
    editor?.contentDOM.blur();
  }

  function updateEditorContents(newContent: string) {
    if (editor && !editor.hasFocus) {
      // NOTE: when changing files, we still want to update the editor
      let curContent = editor.state.doc.toString();
      if (newContent != curContent) {
        editor.dispatch({
          changes: {
            from: 0,
            to: curContent.length,
            insert: newContent,
          },
        });
      }
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      save();
    }
  }

  function save() {
    dispatch("save");
  }

  const debounceSave = debounce(save, FILE_SAVE_DEBOUNCE_TIME);

  function revertContent() {
    editor?.dispatch({
      changes: {
        from: 0,
        to: editor.state.doc.length,
        insert: remoteContent,
      },
    });
    dispatch("revert");
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
        <Button disabled={!hasUnsavedChanges} on:click={save}>
          <Check size="14px" />
          Save
        </Button>

        <Button
          type="text"
          disabled={!hasUnsavedChanges}
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
