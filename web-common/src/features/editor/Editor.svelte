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
  import { FileArtifact } from "../entity-management/file-artifact";
  import Codespace from "./Codespace.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipShortcutContainer from "@rilldata/web-common/components/tooltip/TooltipShortcutContainer.svelte";
  import Shortcut from "@rilldata/web-common/components/tooltip/Shortcut.svelte";
  import MetaKey from "@rilldata/web-common/components/tooltip/MetaKey.svelte";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog/";
  import Alert from "@rilldata/web-common/components/icons/Alert.svelte";

  export let fileArtifact: FileArtifact;
  export let extensions: Extension[] = [];
  export let autoSave = true;
  export let editor: EditorView;
  export let forceLocalUpdates = false;
  export let forceDisableAutoSave = false;
  export let showSaveBar = true;
  export let refetchOnWindowFocus = true;
  export let onSave: (content: string) => void = () => {};
  export let onRevert: () => void = () => {};

  let codespace: Codespace;
  let editorHasFocus = false;

  $: ({
    saveLocalContent,
    revert,
    merging,
    localContent,
    disableAutoSave,
    inConflict,
    saveState: { saving, error, resolve },
    saveEnabled,
  } = fileArtifact);

  $: debounceSave = debounce(save, FILE_SAVE_DEBOUNCE_TIME);

  $: disabled = !$saveEnabled;

  async function handleKeydown(e: KeyboardEvent) {
    if (e.key === "s" && (e.ctrlKey || e.metaKey)) {
      e.preventDefault();
      if (disabled) return;
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
    revert(true); // Revert fileArtifact to remote content
    resolve();
    onRevert(); // Call revert callback
  }

  async function handleRefocus() {
    if (refetchOnWindowFocus) await fileArtifact.fetchContent(true);
  }
</script>

<svelte:window on:keydown={handleKeydown} on:focus={handleRefocus} />

<section>
  {#if $merging}
    <div class="flex w-full border-b">
      <div class="w-full border-r p-1.5 pl-3 flex justify-between items-center">
        <h1 class="text-sm italic font-semibold text-gray-400">
          Unsaved changes
        </h1>

        <Button
          type="subtle"
          loading={$saving}
          loadingCopy="Saving"
          danger={!!$error && !$saving}
          {disabled}
          on:click={async () => {
            await save();
            codespace.mountEditor();
          }}
        >
          {#if $error}
            <Alert size="14px" />
          {/if}

          {#if $error}
            {$error?.message} Try again.
          {:else}
            Accept current
          {/if}
        </Button>
      </div>
      <div class="w-full p-1.5 pl-3 flex justify-between items-center">
        <h2 class="text-sm font-semibold">Incoming content</h2>
        <Button
          type="primary"
          on:click={() => {
            codespace.mountEditor();
            revertContent();
          }}
        >
          Accept incoming
        </Button>
      </div>
    </div>
  {/if}
  <div class="editor-container">
    {#key fileArtifact}
      <Codespace
        {extensions}
        {forceLocalUpdates}
        {fileArtifact}
        autoSave={!forceDisableAutoSave && !disableAutoSave && autoSave}
        bind:editor
        bind:editorHasFocus
        bind:this={codespace}
      />
    {/key}
  </div>

  {#if !$merging && showSaveBar}
    <footer>
      <div class="flex gap-x-3">
        {#if !autoSave || disableAutoSave || forceDisableAutoSave}
          <Tooltip distance={8} activeDelay={300}>
            <Button
              type="subtle"
              loading={$saving}
              danger={!!$error && !$saving}
              loadingCopy="Saving"
              {disabled}
              on:click={save}
            >
              {#if $error}
                <Alert size="14px" />
              {:else}
                <Check size="14px" />
              {/if}

              {#if $error}
                {$error?.message} Try again.
              {:else}
                Save
              {/if}
            </Button>
            <TooltipContent slot="tooltip-content">
              <TooltipShortcutContainer pad={false}>
                Save
                <Shortcut>
                  <MetaKey action="S" />
                </Shortcut>
              </TooltipShortcutContainer>
            </TooltipContent>
          </Tooltip>

          <Button type="text" {disabled} on:click={revertContent}>
            <UndoIcon size="14px" />
            Revert changes
          </Button>
        {/if}
      </div>
      <div
        class="flex gap-x-1 items-center h-full bg-white rounded-full"
        class:hidden={disableAutoSave || forceDisableAutoSave}
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
  {/if}
</section>

{#if $inConflict && !$merging}
  <AlertDialog.Root open>
    <AlertDialog.Content>
      <AlertDialog.Title>File update detected</AlertDialog.Title>
      <AlertDialog.Description>
        This file has been modified by another application. Please resolve
        conflicts with your unsaved changes before proceeding.
      </AlertDialog.Description>

      <AlertDialog.Footer>
        <AlertDialog.Action asChild let:builder>
          <Button
            builders={[builder]}
            type="primary"
            large
            on:click={codespace.mountMergeView}
          >
            Resolve conflicts
          </Button>
        </AlertDialog.Action>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  </AlertDialog.Root>
{/if}

<style lang="postcss">
  .editor-container {
    @apply size-full overflow-auto p-0 pt-0 pb-0 flex flex-col;
  }

  footer {
    @apply justify-between items-center flex flex-none;
    @apply h-10 p-2 w-full rounded-b-sm border-t bg-white;
  }

  section {
    @apply size-full flex-col bg-white flex overflow-hidden relative;
  }
</style>
