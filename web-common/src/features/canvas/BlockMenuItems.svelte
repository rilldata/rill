<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { Copy, SlidersHorizontal, Trash2, Ungroup } from "lucide-svelte";
  import type { BlockKind } from "./stores/block-edit";

  // The block-level items of a dropdown menu, shown for the handlers that are passed:
  // the tab strip offers Duplicate and Add viewer control for its group in each tab's menu,
  // the container chip offers all of them. Add viewer control needs the canvasControls flag.
  let {
    kind,
    onDuplicate,
    onAddControl,
    onUngroup,
    onDelete,
  }: {
    kind: BlockKind;
    onDuplicate?: () => void;
    onAddControl?: () => void;
    onUngroup?: () => void;
    onDelete?: () => void;
  } = $props();

  const { canvasControls } = featureFlags;

  const duplicateLabel = $derived(
    kind === "tab-group"
      ? m.canvas_block_duplicate_tab_group()
      : m.canvas_block_duplicate_container(),
  );
  const deleteLabel = $derived(
    kind === "tab-group"
      ? m.canvas_block_delete_tab_group()
      : m.canvas_block_delete_container(),
  );
</script>

{#if onDuplicate}
  <DropdownMenu.Item
    class="flex flex-row gap-x-2 text-fg-primary"
    onclick={onDuplicate}
  >
    <Copy size="14px" />
    {duplicateLabel}
  </DropdownMenu.Item>
{/if}
{#if onAddControl && $canvasControls}
  <DropdownMenu.Item
    class="flex flex-row gap-x-2 text-fg-primary"
    onclick={onAddControl}
  >
    <SlidersHorizontal size="14px" />
    {m.canvas_viewer_controls_add_viewer()}
  </DropdownMenu.Item>
{/if}
{#if onUngroup}
  <DropdownMenu.Item
    class="flex flex-row gap-x-2 text-fg-primary"
    onclick={onUngroup}
  >
    <Ungroup size="14px" />
    {m.canvas_container_ungroup()}
  </DropdownMenu.Item>
{/if}
{#if onDelete}
  <DropdownMenu.Separator />
  <DropdownMenu.Item
    class="flex flex-row gap-x-2 text-red-600 data-[highlighted]:text-red-600"
    onclick={onDelete}
  >
    <Trash2 size="14px" />
    {deleteLabel}
  </DropdownMenu.Item>
{/if}
