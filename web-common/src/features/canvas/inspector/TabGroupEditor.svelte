<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { Trash2 } from "lucide-svelte";
  import { parseDocument, type Document } from "yaml";
  import TabListItem from "./TabListItem.svelte";
  import type { TabGroup } from "../stores/tab-group";
  import {
    deleteTab,
    deleteTabGroup,
    duplicateTab,
    moveTab,
    renameTab,
    renameTabGroup,
    setTabName,
    tabHasContent,
  } from "../stores/tab-edit";

  export let group: TabGroup;
  export let blockIndex: number;
  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;
  // Clear the tab-group selection (e.g. after deleting the whole group).
  export let onClose: () => void;
  // Re-point the selection at the group's new name after a rename, so the inspector stays
  // open on it (the group is re-keyed by name when the spec reprocesses).
  export let onRename: (name: string) => void;

  $: ({ editorContent, updateEditorContent, saveLocalContent } = fileArtifact);

  $: tabs = group.tabs;
  $: activeTabIndex = group.activeTabIndex;
  $: tabCount = $tabs.length;

  // Local copy of the group name. The `group` prop is a fresh object only when the group is
  // re-keyed (i.e. renamed), so syncing on group.name keeps typing from being clobbered.
  let groupName = "";
  let lastGroupName = "";
  $: if (group.name !== lastGroupName) {
    groupName = group.name;
    lastGroupName = group.name;
  }

  // Index pending a delete confirmation, or null when no dialog is open.
  let pendingTabDelete: number | null = null;
  let pendingGroupDelete = false;

  async function applyEdit(mutate: (doc: Document) => void) {
    const doc = parseDocument($editorContent ?? "");
    mutate(doc);
    updateEditorContent(doc.toString(), false, autoSave);
    await saveLocalContent();
  }

  async function commitGroupRename() {
    const trimmed = groupName.trim();
    if (trimmed === group.name) return;
    await applyEdit((doc) => renameTabGroup(doc, blockIndex, trimmed));
    // Follow the group to its new key (an empty name defaults to `group-<index>`).
    onRename(trimmed || `group-${blockIndex}`);
  }

  async function commitLabel(index: number, value: string) {
    const trimmed = value.trim();
    if (!trimmed) return;
    // Compare against the persisted YAML, not $tabs[index].displayName: the live typing
    // optimistically updates the latter (so the strip reflects edits), which would make a
    // naive guard think nothing changed and skip the save, losing the edit on refresh.
    const persisted = parseDocument($editorContent ?? "").getIn([
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
    // Keep the moved tab active by name: its destination index is only known after the spec
    // reflects the reorder, and matching by name survives the index shuffle.
    const movedName = $tabs[index]?.name;
    if (movedName) group.activateByNameWhenReady(movedName);
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
    if (tabHasContent(parseDocument($editorContent ?? ""), blockIndex, index)) {
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
    if (wasLastTab) onClose();
  }

  function requestDeleteGroup() {
    const doc = parseDocument($editorContent ?? "");
    const hasContent = $tabs.some((_, i) => tabHasContent(doc, blockIndex, i));
    if (hasContent) {
      pendingGroupDelete = true;
    } else {
      void confirmDeleteGroup();
    }
  }

  async function confirmDeleteGroup() {
    pendingGroupDelete = false;
    await applyEdit((doc) => {
      deleteTabGroup(doc, blockIndex);
    });
    onClose();
  }
</script>

<SidebarWrapper type="secondary" disableHorizontalPadding title="Tab group">
  <div class="param">
    <Input
      capitalizeLabel={false}
      textClass="text-sm"
      size="sm"
      labelGap={2}
      label="Tab group name"
      hint="Stable identifier used as the group's deep-link URL key"
      bind:value={groupName}
      onBlur={commitGroupRename}
      onEnter={commitGroupRename}
    />
  </div>

  <div class="param flex flex-col gap-y-2">
    <span class="text-xs font-medium text-fg-secondary">Tabs</span>
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

  <div class="param">
    <Button type="secondary-destructive" onClick={requestDeleteGroup}>
      <Trash2 size="14px" />
      Delete tab group
    </Button>
  </div>
</SidebarWrapper>

{#if pendingTabDelete !== null}
  {@const index = pendingTabDelete}
  <AlertDialog.Root
    open
    onOpenChange={(open) => !open && (pendingTabDelete = null)}
  >
    <AlertDialog.Content>
      <AlertDialog.Title>Delete tab?</AlertDialog.Title>
      <AlertDialog.Description>
        This tab and all of its widgets will be permanently removed.
      </AlertDialog.Description>
      <AlertDialog.Footer>
        <AlertDialog.Cancel>
          {#snippet child({ props })}
            <Button {...props} large type="secondary">Cancel</Button>
          {/snippet}
        </AlertDialog.Cancel>
        <AlertDialog.Action>
          {#snippet child({ props })}
            <Button
              {...props}
              large
              type="destructive"
              onClick={() => confirmDeleteTab(index)}
            >
              Delete
            </Button>
          {/snippet}
        </AlertDialog.Action>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  </AlertDialog.Root>
{/if}

{#if pendingGroupDelete}
  <AlertDialog.Root open onOpenChange={(open) => (pendingGroupDelete = open)}>
    <AlertDialog.Content>
      <AlertDialog.Title>Delete tab group?</AlertDialog.Title>
      <AlertDialog.Description>
        This tab group and all of its tabs and widgets will be permanently
        removed.
      </AlertDialog.Description>
      <AlertDialog.Footer>
        <AlertDialog.Cancel>
          {#snippet child({ props })}
            <Button {...props} large type="secondary">Cancel</Button>
          {/snippet}
        </AlertDialog.Cancel>
        <AlertDialog.Action>
          {#snippet child({ props })}
            <Button
              {...props}
              large
              type="destructive"
              onClick={confirmDeleteGroup}
            >
              Delete
            </Button>
          {/snippet}
        </AlertDialog.Action>
      </AlertDialog.Footer>
    </AlertDialog.Content>
  </AlertDialog.Root>
{/if}

<style lang="postcss">
  .param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
