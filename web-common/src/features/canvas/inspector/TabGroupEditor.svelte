<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { get } from "svelte/store";
  import { parseDocument, type Document } from "yaml";
  import type { CanvasEntity } from "../stores/canvas-entity";
  import {
    deleteTab,
    duplicateTab,
    moveTab,
    renameTab,
    setTabName,
    tabHasContent,
  } from "../stores/tab-edit";
  import type { TabGroup } from "../stores/tab-group";
  import BlockEditor, { applyCanvasEdit } from "./BlockEditor.svelte";
  import TabListItem from "./TabListItem.svelte";

  // The inspector of a tab group: the shared block inspector with the group's tabs.
  let {
    group,
    blockIndex,
    canvas,
    fileArtifact,
    autoSave,
  }: {
    group: TabGroup;
    blockIndex: number;
    canvas: CanvasEntity;
    fileArtifact: FileArtifact;
    autoSave: boolean;
  } = $props();

  const tabs = $derived(group.tabs);
  const activeTabIndex = $derived(group.activeTabIndex);
  const tabCount = $derived($tabs.length);

  // Index pending a delete confirmation, or null when no dialog is open.
  let pendingTabDelete = $state<number | null>(null);

  function currentDoc() {
    return parseDocument(get(fileArtifact.editorContent) ?? "");
  }

  function applyEdit(mutate: (doc: Document) => void) {
    return applyCanvasEdit(fileArtifact, autoSave, mutate);
  }

  async function commitLabel(index: number, value: string) {
    const trimmed = value.trim();
    if (!trimmed) return;
    // Compare against the persisted YAML, not $tabs[index].displayName: the live typing
    // optimistically updates the latter (so the strip reflects edits), which would make a
    // naive guard think nothing changed and skip the save, losing the edit on refresh.
    const persisted = currentDoc().getIn([
      "rows",
      blockIndex,
      "tabs",
      index,
      "label",
    ]);
    if (trimmed === persisted) return;
    await applyEdit((doc) => renameTab(doc, blockIndex, index, trimmed));
  }

  async function commitName(index: number, value: string) {
    if (value.trim() === $tabs[index]?.name) return;
    await applyEdit((doc) => setTabName(doc, blockIndex, index, value));
  }

  async function move(index: number, direction: -1 | 1) {
    // The active tab is preserved by name across the reorder (see TabGroup.updateFromSpec).
    await applyEdit((doc) => moveTab(doc, blockIndex, index, direction));
  }

  async function duplicate(index: number) {
    // duplicateTab inserts the copy immediately after the original; activate it.
    group.activateWhenReady(index + 1);
    await applyEdit((doc) => {
      duplicateTab(doc, blockIndex, index);
    });
  }

  function requestDeleteTab(index: number) {
    if (tabHasContent(currentDoc(), blockIndex, index)) {
      pendingTabDelete = index;
    } else {
      void confirmDeleteTab(index);
    }
  }

  async function confirmDeleteTab(index: number) {
    pendingTabDelete = null;
    const wasLastTab = tabCount <= 1;
    await applyEdit((doc) => deleteTab(doc, blockIndex, index));
    // Deleting the last tab unwraps the group, so the group no longer exists.
    if (wasLastTab) canvas.setSelectedBlock(null);
  }
</script>

<BlockEditor
  kind="tab-group"
  name={group.name}
  {blockIndex}
  {canvas}
  {fileArtifact}
  {autoSave}
>
  <div class="param flex flex-col gap-y-2">
    <span class="text-xs font-medium text-fg-secondary">
      {m.canvas_tabs()}
    </span>
    <ul class="flex flex-col gap-y-2">
      {#each $tabs as tab, index (index)}
        <TabListItem
          displayName={tab.displayName}
          name={tab.name}
          active={index === $activeTabIndex}
          canMoveUp={index > 0}
          canMoveDown={index < tabCount - 1}
          onCommitLabel={(value) => commitLabel(index, value)}
          onInputLabel={(value) => group.setTabDisplayName(index, value)}
          onCommitName={(value) => commitName(index, value)}
          onMoveUp={() => move(index, -1)}
          onMoveDown={() => move(index, 1)}
          onDuplicate={() => duplicate(index)}
          onDelete={() => requestDeleteTab(index)}
        />
      {/each}
    </ul>
  </div>
</BlockEditor>

<AlertDialog.Confirmation
  open={pendingTabDelete !== null}
  onOpenChange={(open) => {
    if (!open) pendingTabDelete = null;
  }}
  title={m.canvas_delete_tab()}
  description={m.canvas_delete_tab_description()}
  confirmLabel={m.common_delete()}
  confirmType="destructive"
  onConfirm={() => {
    if (pendingTabDelete !== null) void confirmDeleteTab(pendingTabDelete);
  }}
/>

<style lang="postcss">
  .param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
