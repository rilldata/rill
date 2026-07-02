<script lang="ts">
  import type { DimensionTag } from "@rilldata/web-common/components/menu/tag-utils";
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Pivot from "@rilldata/web-common/components/icons/Pivot.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { detectOverflow } from "@rilldata/web-common/lib/actions/detect-overflow";
  import { modifierHeld } from "@rilldata/web-common/lib/modifier-key";
  import { dragDataStore } from "./DragList.svelte";
  import PivotPortalItem from "./PivotPortalItem.svelte";
  import { appendChipsToZone, replaceZoneCleaningOther } from "./pivot-utils";
  import { PivotChipType, type PivotChipData } from "./types";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  type Props = {
    tag: DimensionTag;
    dimensions: PivotChipData[];
    measures: PivotChipData[];
    rows: PivotChipData[];
    columns: PivotChipData[];
    selected: boolean;
    setRows: (items: PivotChipData[]) => void;
    setColumns: (items: PivotChipData[]) => void;
    onSelect: () => void;
  };

  let {
    tag,
    dimensions,
    measures,
    rows,
    columns,
    selected,
    setRows,
    setColumns,
    onSelect,
  }: Props = $props();

  const DRAG_THRESHOLD_PX = 4;

  let rowEl: HTMLDivElement | undefined = $state();
  let pendingDrag: {
    rect: DOMRect;
    startX: number;
    startY: number;
    offsetX: number;
    offsetY: number;
  } | null = null;
  let dragActive = $state(false);
  let dragPosition = $state({ left: 0, top: 0 });
  let dragOffset = $state({ x: 0, y: 0 });
  let dragChip: PivotChipData | null = $state(null);

  const dimensionCount = $derived(dimensions.length);
  const measureCount = $derived(measures.length);

  let isTruncated = $state(false);

  const actionBtnClass =
    "flex items-center justify-center h-[18px] w-[18px] rounded-sm text-icon-muted hover:text-fg-primary hover:bg-surface-background transition-colors";

  // ---- mutation actions ---- //

  function addToRows(replace: boolean) {
    if (replace) {
      const { zone, other } = replaceZoneCleaningOther(dimensions, columns);
      setRows(zone);
      if (other.length !== columns.length) setColumns(other);
      return;
    }
    if (dimensions.length === 0) return;
    setRows(appendChipsToZone(rows, columns, dimensions));
  }

  function addToColumns(replace: boolean) {
    const all = [...dimensions, ...measures];
    if (replace) {
      const { zone, other } = replaceZoneCleaningOther(all, rows);
      setColumns(zone);
      if (other.length !== rows.length) setRows(other);
      return;
    }
    if (all.length === 0) return;
    setColumns(appendChipsToZone(columns, rows, all));
  }

  function autoArrange(replace: boolean) {
    if (replace) {
      // Replace both zones; cross-zone cleanup happens naturally since each
      // zone gets exactly its slice of the tag.
      setRows(dimensions);
      setColumns(measures);
      return;
    }
    if (dimensions.length > 0) {
      setRows(appendChipsToZone(rows, columns, dimensions));
    }
    if (measures.length > 0) {
      setColumns(appendChipsToZone(columns, rows, measures));
    }
  }

  // ---- click vs drag ---- //

  function handleMouseDown(e: MouseEvent) {
    if (e.button !== 0) return;
    const target = e.target as HTMLElement | null;
    if (target?.closest("button")) return;
    if (!rowEl) return;
    e.preventDefault();
    const rect = rowEl.getBoundingClientRect();
    pendingDrag = {
      rect,
      startX: e.clientX,
      startY: e.clientY,
      offsetX: e.clientX - rect.left,
      offsetY: e.clientY - rect.top,
    };
    window.addEventListener("mousemove", detectDragStart);
    window.addEventListener("mouseup", handleGlobalMouseUp, { once: true });
  }

  function detectDragStart(e: MouseEvent) {
    if (!pendingDrag || dragActive) return;
    const moved =
      Math.abs(e.clientX - pendingDrag.startX) >= DRAG_THRESHOLD_PX ||
      Math.abs(e.clientY - pendingDrag.startY) >= DRAG_THRESHOLD_PX;
    if (!moved) return;
    beginDrag();
  }

  function beginDrag() {
    if (!pendingDrag) return;
    dragActive = true;
    window.removeEventListener("mousemove", detectDragStart);

    const { rect, offsetX, offsetY } = pendingDrag;
    dragPosition = { left: rect.left, top: rect.top };
    dragOffset = { x: offsetX, y: offsetY };

    // Synthetic chip used by PivotPortalItem to render the floating preview.
    // Pure-measure tags render with the rectangular measure chip; mixed and
    // pure-dimension tags use the rounded dimension shape.
    const chipType =
      dimensions.length === 0 && measures.length > 0
        ? PivotChipType.Measure
        : PivotChipType.Dimension;
    dragChip = {
      id: `__tag__:${tag.name}`,
      title: tag.name,
      type: chipType,
    };

    dragDataStore.set({
      source: "tags",
      width: rect.width,
      chip: dragChip,
      tagPayload: { tagName: tag.name, dimensions, measures },
    });
  }

  function handleGlobalMouseUp() {
    window.removeEventListener("mousemove", detectDragStart);
    if (!dragActive) {
      // Mousedown without movement: treat as a click on the tag row body.
      // Toggles the filter selection.
      if (pendingDrag) onSelect();
      pendingDrag = null;
      return;
    }
    resetDrag();
  }

  function resetDrag() {
    dragActive = false;
    dragChip = null;
    pendingDrag = null;
    dragDataStore.set(null);
    window.removeEventListener("mousemove", detectDragStart);
  }

  function handleActionClick(
    e: MouseEvent,
    action: (replace: boolean) => void,
  ) {
    action(e.metaKey || e.ctrlKey);
  }
</script>

<div
  bind:this={rowEl}
  class="tag-row group"
  class:selected
  role="presentation"
  onmousedown={handleMouseDown}
>
  <Tooltip.Root delayDuration={200} disabled={!isTruncated}>
    <Tooltip.Trigger>
      {#snippet child({ props })}
        <span
          {...props}
          class="truncate flex-1 min-w-0 text-left text-fg-primary"
          use:detectOverflow={(v) => (isTruncated = v)}
        >
          {tag.name}
        </span>
      {/snippet}
    </Tooltip.Trigger>
    <Tooltip.Content
      side="top"
      class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
    >
      {tag.name}
    </Tooltip.Content>
  </Tooltip.Root>

  <div class="flex items-center gap-x-1 flex-none group-hover:hidden">
    {#if measureCount > 0}
      <span
        class="count-tile meas-tile"
        title={m.dashboard_measures_count({ count: measureCount.toString() })}
      >
        {measureCount}
      </span>
    {/if}
    {#if dimensionCount > 0}
      <span
        class="count-tile dim-tile"
        title={m.dashboard_dimensions_count({
          count: dimensionCount.toString(),
        })}
      >
        {dimensionCount}
      </span>
    {/if}
  </div>

  <div class="hidden group-hover:flex items-center gap-x-0.5 flex-none">
    {#if dimensionCount > 0}
      <Tooltip.Root delayDuration={200}>
        <Tooltip.Trigger>
          {#snippet child({ props })}
            <button
              {...props}
              type="button"
              class={actionBtnClass}
              onclick={(e) => handleActionClick(e, addToRows)}
              aria-label={$modifierHeld
                ? m.dashboard_replace_rows_tag({ name: tag.name })
                : m.dashboard_add_all_dims_rows({ name: tag.name })}
            >
              <Row size="14px" color="currentColor" />
            </button>
          {/snippet}
        </Tooltip.Trigger>
        <Tooltip.Content
          side="top"
          class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
        >
          {#if $modifierHeld}
            {m.dashboard_replace_rows_tag_dims()}
          {:else}
            <div>{m.dashboard_add_all_to_rows()}</div>
            <div class="hint">{m.dashboard_cmd_click_replace()}</div>
          {/if}
        </Tooltip.Content>
      </Tooltip.Root>
    {/if}

    <Tooltip.Root delayDuration={200}>
      <Tooltip.Trigger>
        {#snippet child({ props })}
          <button
            {...props}
            type="button"
            class={actionBtnClass}
            onclick={(e) => handleActionClick(e, addToColumns)}
            aria-label={$modifierHeld
              ? m.dashboard_replace_columns_tag({ name: tag.name })
              : m.dashboard_add_all_to_columns_tag({ name: tag.name })}
          >
            <Column size="14px" color="currentColor" />
          </button>
        {/snippet}
      </Tooltip.Trigger>
      <Tooltip.Content
        side="top"
        class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
      >
        {#if $modifierHeld}
          {m.dashboard_replace_columns_tag_items()}
        {:else}
          <div>{m.dashboard_add_all_to_columns()}</div>
          <div class="hint">{m.dashboard_cmd_click_replace()}</div>
        {/if}
      </Tooltip.Content>
    </Tooltip.Root>

    {#if dimensionCount > 0 && measureCount > 0}
      <Tooltip.Root delayDuration={200}>
        <Tooltip.Trigger>
          {#snippet child({ props })}
            <button
              {...props}
              type="button"
              class={actionBtnClass}
              onclick={(e) => handleActionClick(e, autoArrange)}
              aria-label={$modifierHeld
                ? m.dashboard_replace_auto_arrange({ name: tag.name })
                : m.dashboard_auto_arrange_tag({ name: tag.name })}
            >
              <Pivot size="14px" color="currentColor" />
            </button>
          {/snippet}
        </Tooltip.Trigger>
        <Tooltip.Content
          side="top"
          class="bg-popover text-fg-primary z-popover text-xs px-2 py-1"
        >
          {#if $modifierHeld}
            {m.dashboard_replace_rows_cols_tag()}
          {:else}
            <div>{m.dashboard_auto_arrange()}</div>
            <div class="hint">{m.dashboard_cmd_click_replace()}</div>
          {/if}
        </Tooltip.Content>
      </Tooltip.Root>
    {/if}
  </div>
</div>

{#if dragActive && dragChip}
  <PivotPortalItem
    item={dragChip}
    offset={dragOffset}
    position={dragPosition}
    removable={false}
    onRelease={resetDrag}
  />
{/if}

<style lang="postcss">
  .tag-row {
    @apply w-full flex items-center gap-x-1 px-1.5 py-0.5 rounded-sm;
    @apply cursor-grab select-none;
    @apply hover:bg-surface-hover;
  }

  .tag-row.selected {
    @apply bg-popover-accent;
  }

  .tag-row.selected:hover {
    @apply bg-popover-accent;
  }

  .tag-row:active {
    @apply cursor-grabbing;
  }

  .count-tile {
    @apply inline-flex items-center justify-center;
    @apply tabular-nums text-[10px] font-medium;
    @apply min-w-[16px] h-[16px] px-1 rounded-sm border;
  }

  .dim-tile {
    @apply bg-theme-50 border-theme-200 text-theme-800;
  }

  .meas-tile {
    @apply bg-theme-secondary-50 border-theme-secondary-200 text-theme-secondary-800;
  }

  .kbd {
    @apply inline-block px-1 py-px rounded-sm border;
    @apply text-[10px] font-mono text-fg-secondary;
  }

  .hint {
    @apply mt-0.5 text-[10px] text-fg-secondary;
  }
</style>
