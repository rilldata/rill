<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Label from "@rilldata/web-common/components/forms/Label.svelte";
  import Switch from "@rilldata/web-common/components/forms/Switch.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import UndoIcon from "@rilldata/web-common/components/icons/UndoIcon.svelte";
  import type { Extension } from "@codemirror/state";

  import { EditorView } from "@codemirror/view";
  import { debounce } from "../../lib/create-debouncer";
  import { FILE_SAVE_DEBOUNCE_TIME } from "./config";
  import { FileArtifact } from "../entity-management/file-artifacts";
  import Codespace from "./Codespace.svelte";

  export let fileArtifact: FileArtifact;
  export let extensions: Extension[] = [];
  export let autoSave: boolean = true;
  export let disableAutoSave: boolean = false;
  export let editor: EditorView | null = null;
  export let forceLocalUpdates: boolean = false;
  export let onSave: (content: string) => void = () => {};
  export let onRevert: () => void = () => {};

  $: ({ hasUnsavedChanges, saveLocalContent, revert, localContent } =
    fileArtifact);

  $: debounceSave = debounce(save, FILE_SAVE_DEBOUNCE_TIME);

  async function handleKeydown(e: KeyboardEvent) {
    if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      await save();
    }
  }

  async function save() {
    const local = $localContent;
    if (local === null) return;
    onSave(local);
    await saveLocalContent();
  }

  function revertContent() {
    revert(); // Revert fileArtifact to remote content
    onRevert(); // Call revert callback
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<section>
  <div class="editor-container">
    {#key fileArtifact}
      <Codespace
        {extensions}
        {debounceSave}
        {forceLocalUpdates}
        {fileArtifact}
        {autoSave}
        {disableAutoSave}
        bind:editor
      />
    {/key}
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
