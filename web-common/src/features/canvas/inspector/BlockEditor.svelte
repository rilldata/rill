<script lang="ts" module>
  import type { FileArtifact } from "@rilldata/web-common/features/entity-management/file-artifact";
  import { get } from "svelte/store";
  import { parseDocument, type Document } from "yaml";
  import { nameBlocks } from "../stores/block-edit";

  // Applies an edit to the canvas YAML and saves it. Blocks are named before the edit, so an
  // edit that shifts rows keeps their keys (see nameBlocks).
  export async function applyCanvasEdit(
    fileArtifact: FileArtifact,
    autoSave: boolean,
    mutate: (doc: Document) => void,
  ) {
    const doc = nameBlocks(
      parseDocument(get(fileArtifact.editorContent) ?? ""),
    );
    mutate(doc);
    fileArtifact.updateEditorContent(doc.toString(), false, autoSave);
    await fileArtifact.saveLocalContent();
  }
</script>

<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SidebarWrapper from "@rilldata/web-common/features/visual-editing/SidebarWrapper.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Copy, Trash2 } from "lucide-svelte";
  import type { Snippet } from "svelte";
  import BlockDeleteDialog from "../BlockDeleteDialog.svelte";
  import {
    blockHasContent,
    blockIndexOf,
    deleteBlock,
    duplicateBlock,
    renameBlock,
    type BlockKind,
  } from "../stores/block-edit";
  import type { CanvasEntity } from "../stores/canvas-entity";
  import VisibilityInput from "./VisibilityInput.svelte";

  // The inspector of a block (a tab group or a container): its name, who sees it, and the
  // actions on the whole block. Kind-specific content, such as a tab group's tabs, goes in
  // `children`; kind-specific actions, such as ungrouping a container, go in `actions`.
  let {
    kind,
    name,
    blockIndex,
    canvas,
    fileArtifact,
    autoSave,
    children,
    actions,
  }: {
    kind: BlockKind;
    // The block's name, which is its key; an unnamed block in the YAML has its default name.
    name: string;
    blockIndex: number;
    canvas: CanvasEntity;
    fileArtifact: FileArtifact;
    autoSave: boolean;
    children?: Snippet;
    actions?: Snippet;
  } = $props();

  // A local draft of the name. It resets when the block is renamed, so typing isn't
  // clobbered by unrelated updates.
  let draftName = $derived(name);
  // A name the author tried that another block already has. Its error shows while the draft
  // still holds it.
  let takenName = $state<string>();
  const nameError = $derived(
    takenName !== undefined &&
      takenName !== name &&
      draftName.trim() === takenName
      ? m.canvas_block_name_taken({ name: takenName })
      : undefined,
  );
  let pendingDelete = $state(false);

  const labels = $derived(
    kind === "tab-group"
      ? {
          title: m.canvas_tab_group(),
          name: m.canvas_block_name_tab_group(),
          nameHint: m.canvas_block_name_hint_tab_group(),
          visibilityHint: m.canvas_visibility_hint_tab_group(),
          duplicate: m.canvas_block_duplicate_tab_group(),
          delete: m.canvas_block_delete_tab_group(),
        }
      : {
          title: m.canvas_container(),
          name: m.canvas_block_name_container(),
          nameHint: m.canvas_block_name_hint_container(),
          visibilityHint: m.canvas_visibility_hint_container(),
          duplicate: m.canvas_block_duplicate_container(),
          delete: m.canvas_block_delete_container(),
        },
  );

  async function commitRename() {
    const trimmed = draftName.trim();
    // Every block has a name, so clearing the field puts the current one back.
    if (!trimmed || trimmed === name) {
      draftName = name;
      return;
    }
    // Tab groups and containers share one set of names, and the parser rejects a canvas in
    // which two blocks have the same one.
    const current = parseDocument(get(fileArtifact.editorContent) ?? "");
    if (blockIndexOf(current, trimmed) !== -1) {
      takenName = trimmed;
      return;
    }
    await applyCanvasEdit(fileArtifact, autoSave, (doc) => {
      // The block's instance and the selection follow it to its new key.
      if (renameBlock(doc, blockIndex, trimmed)) {
        canvas.renameBlockKey(name, trimmed);
      }
    });
  }

  async function duplicate() {
    await applyCanvasEdit(fileArtifact, autoSave, (doc) => {
      const copy = duplicateBlock(doc, blockIndex);
      if (copy) canvas.setSelectedBlock(copy);
    });
  }

  function requestDelete() {
    const doc = parseDocument(get(fileArtifact.editorContent) ?? "");
    if (blockHasContent(doc, blockIndex)) {
      pendingDelete = true;
    } else {
      void confirmDelete();
    }
  }

  async function confirmDelete() {
    pendingDelete = false;
    await applyCanvasEdit(fileArtifact, autoSave, (doc) => {
      if (deleteBlock(doc, blockIndex)) canvas.setSelectedBlock(null);
    });
  }
</script>

<SidebarWrapper type="secondary" disableHorizontalPadding title={labels.title}>
  <div class="param">
    <Input
      id="canvas-block-name"
      capitalizeLabel={false}
      textClass="text-sm"
      size="sm"
      labelGap={2}
      label={labels.name}
      hint={labels.nameHint}
      errors={nameError}
      alwaysShowError
      bind:value={draftName}
      onBlur={commitRename}
      onEnter={commitRename}
    />
  </div>

  <VisibilityInput
    {canvas}
    path={["rows", blockIndex]}
    hint={labels.visibilityHint}
    dataNote={m.canvas_visibility_block_data_note()}
    id="canvas-block-visibility-condition"
  />

  {@render children?.()}

  <div class="param flex flex-col gap-y-2">
    <Button type="secondary" onClick={duplicate}>
      <Copy size="14px" />
      {labels.duplicate}
    </Button>
    {@render actions?.()}
    <Button type="secondary-destructive" onClick={requestDelete}>
      <Trash2 size="14px" />
      {labels.delete}
    </Button>
  </div>
</SidebarWrapper>

{#if pendingDelete}
  <BlockDeleteDialog
    {kind}
    onConfirm={confirmDelete}
    onCancel={() => (pendingDelete = false)}
  />
{/if}

<style lang="postcss">
  .param {
    @apply py-3 px-5;
    @apply border-t;
  }
</style>
