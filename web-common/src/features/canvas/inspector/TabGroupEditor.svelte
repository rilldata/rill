<script lang="ts">
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { ArrowLeft, ArrowRight, Copy, Trash2 } from "lucide-svelte";
  import { parseDocument, type Document } from "yaml";
  import type { TabGroup } from "../stores/tab-group";
  import {
    deleteTab,
    deleteTabGroup,
    duplicateTab,
    moveTab,
    renameTab,
    tabHasContent,
  } from "../stores/tab-edit";

  export let group: TabGroup;
  export let blockIndex: number;
  export let fileArtifact: FileArtifact;
  export let autoSave: boolean;
  // Clear the tab-group selection (e.g. after deleting the whole group).
  export let onClose: () => void;

  $: ({ editorContent, updateEditorContent, saveLocalContent } = fileArtifact);

  $: tabs = group.tabs;
  $: activeTabIndex = group.activeTabIndex;
  $: activeTab = $tabs[$activeTabIndex];
  $: tabCount = $tabs.length;

  // Local copy of the active tab's label for inline editing, reset only when the active
  // tab's identity changes (not on every reconcile) so typing isn't clobbered.
  let label = "";
  let labelTabName = "";
  $: if (activeTab && activeTab.name !== labelTabName) {
    label = activeTab.displayName;
    labelTabName = activeTab.name;
  }

  let pendingTabDelete = false;
  let pendingGroupDelete = false;

  async function applyEdit(mutate: (doc: Document) => void) {
    const doc = parseDocument($editorContent ?? "");
    mutate(doc);
    updateEditorContent(doc.toString(), false, autoSave);
    await saveLocalContent();
  }

  async function commitRename() {
    const trimmed = label.trim();
    if (!trimmed || trimmed === activeTab?.displayName) return;
    await applyEdit((doc) =>
      renameTab(doc, blockIndex, $activeTabIndex, trimmed),
    );
  }

  async function move(direction: -1 | 1) {
    const target = $activeTabIndex + direction;
    await applyEdit((doc) =>
      moveTab(doc, blockIndex, $activeTabIndex, direction),
    );
    group.activateWhenReady(target);
  }

  async function duplicate() {
    const index = $activeTabIndex;
    let newIndex = -1;
    await applyEdit((doc) => {
      newIndex = duplicateTab(doc, blockIndex, index);
    });
    if (newIndex >= 0) group.activateWhenReady(newIndex);
  }

  function requestDeleteTab() {
    if (
      tabHasContent(
        parseDocument($editorContent ?? ""),
        blockIndex,
        $activeTabIndex,
      )
    ) {
      pendingTabDelete = true;
    } else {
      void confirmDeleteTab();
    }
  }

  async function confirmDeleteTab() {
    pendingTabDelete = false;
    const wasLastTab = tabCount <= 1;
    await applyEdit((doc) => deleteTab(doc, blockIndex, $activeTabIndex));
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
      size="sm"
      labelGap={2}
      label="Tab name"
      bind:value={label}
      onBlur={commitRename}
      onEnter={commitRename}
    />
  </div>

  <div class="param flex flex-col gap-y-2">
    <span class="text-xs font-medium text-fg-secondary">Tab</span>
    <div class="flex flex-wrap gap-2">
      <Button
        type="secondary"
        small
        disabled={$activeTabIndex === 0}
        onClick={() => move(-1)}
      >
        <ArrowLeft size="14px" />
        Move left
      </Button>
      <Button
        type="secondary"
        small
        disabled={$activeTabIndex >= tabCount - 1}
        onClick={() => move(1)}
      >
        <ArrowRight size="14px" />
        Move right
      </Button>
      <Button type="secondary" small onClick={duplicate}>
        <Copy size="14px" />
        Duplicate
      </Button>
      <Button type="secondary" small onClick={requestDeleteTab}>
        <Trash2 size="14px" />
        Delete tab
      </Button>
    </div>
  </div>

  <div class="param">
    <Button type="secondary-destructive" onClick={requestDeleteGroup}>
      <Trash2 size="14px" />
      Delete tab group
    </Button>
  </div>
</SidebarWrapper>

{#if pendingTabDelete}
  <AlertDialog.Root open onOpenChange={(open) => (pendingTabDelete = open)}>
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
              onClick={confirmDeleteTab}
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
