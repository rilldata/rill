<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import MeasureFormatChip from "@rilldata/web-common/features/dashboards/pivot/MeasureFormatChip.svelte";
  import PivotPortalItem from "@rilldata/web-common/features/dashboards/pivot/PivotPortalItem.svelte";
  import { swapListener } from "@rilldata/web-common/features/dashboards/pivot/swapListener";
  import {
    PivotChipType,
    type PivotChipData,
    type PivotMeasureFormatting,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { writable } from "svelte/store";
  import type { FieldType } from "./types";

  const _ghostIndex = writable<number | null>(null);

  export let items: string[] = [];
  export let displayMap: Record<string, { label: string; type: FieldType }>;
  export let onUpdate: (items: string[]) => void;
  export let orientation: "horizontal" | "vertical" = "vertical";
  // When provided, measure chips expose per-measure conditional formatting
  // controls in a dropdown on the chip.
  export let measureFormatting:
    | Record<string, PivotMeasureFormatting>
    | undefined = undefined;
  export let setMeasureFormatting:
    | ((measureName: string, fmt: PivotMeasureFormatting | null) => void)
    | undefined = undefined;
  export let lowerIsBetterMap: Record<string, boolean> = {};

  // Distinguish click (open the format dropdown) from drag (reorder): the drag
  // only begins once the pointer moves past this threshold.
  const DRAG_START_THRESHOLD_PX = 4;

  type PendingDragState = {
    item: string;
    index: number;
    width: number;
    left: number;
    top: number;
    offsetX: number;
    offsetY: number;
    startX: number;
    startY: number;
  };

  let pendingDrag: PendingDragState | null = null;
  let dragData: PivotChipData | null = null;
  let dragStart = { left: 0, top: 0 };
  let offset = { x: 0, y: 0 };
  let draggedItemWidth = 0;

  $: ghostIndex = $_ghostIndex;

  function toChipData(item: string): PivotChipData {
    return {
      id: item,
      title: displayMap[item]?.label?.replace(/Time /, "") || item,
      type:
        displayMap[item]?.type === "measure"
          ? PivotChipType.Measure
          : displayMap[item]?.type === "time"
            ? PivotChipType.Time
            : PivotChipType.Dimension,
    };
  }

  function handleMouseDown(e: MouseEvent, item: string, index: number) {
    // No-op if the target is a chip remove button or the format dropdown
    const target = e.target as HTMLElement;
    if (
      target.closest('[aria-label="Remove"]') ||
      target.closest(".format-dropdown")
    )
      return;

    if (e.button !== 0) return;

    e.preventDefault();
    const dragItem = e.currentTarget as HTMLElement;
    if (!dragItem) return;

    const { width, left, top } = dragItem.getBoundingClientRect();

    pendingDrag = {
      item,
      index,
      width,
      left,
      top,
      offsetX: e.clientX - left,
      offsetY: e.clientY - top,
      startX: e.clientX,
      startY: e.clientY,
    };

    window.addEventListener("mousemove", detectDragStart);
    window.addEventListener("mouseup", handleMouseUp, { once: true });
  }

  function detectDragStart(e: MouseEvent) {
    if (!pendingDrag) return;

    const movedBeyondThreshold =
      Math.abs(e.clientX - pendingDrag.startX) >= DRAG_START_THRESHOLD_PX ||
      Math.abs(e.clientY - pendingDrag.startY) >= DRAG_START_THRESHOLD_PX;

    if (!movedBeyondThreshold) return;

    beginDrag();
  }

  function beginDrag() {
    if (!pendingDrag) return;

    window.removeEventListener("mousemove", detectDragStart);

    const { item, index, width, left, top, offsetX, offsetY } = pendingDrag;
    pendingDrag = null;

    dragStart = { left, top };
    offset = { x: offsetX, y: offsetY };

    dragData = toChipData(item);
    draggedItemWidth = width;
    _ghostIndex.set(index);

    const temp = [...items];
    temp.splice(index, 1);
    items = temp;
  }

  function handleMouseUp() {
    window.removeEventListener("mousemove", detectDragStart);

    // The pointer never moved past the threshold: treat as a click, not a drag.
    if (pendingDrag) {
      pendingDrag = null;
      return;
    }

    handleDrop();
  }

  function handleDrop() {
    if (dragData && ghostIndex !== null) {
      if (!items.includes(dragData.id)) {
        const temp = [...items];
        temp.splice(ghostIndex, 0, dragData.id);
        items = temp;
        onUpdate(items);
      } else {
        console.warn("Prevented duplicate addition of item:", dragData.title);
      }
    }
    dragData = null;
    _ghostIndex.set(null);
  }

  function handleRemove(item: string) {
    const temp = [...items];
    const index = temp.indexOf(item);
    if (index !== -1) {
      temp.splice(index, 1);
      items = temp;
      onUpdate(items);
    }
  }
</script>

<div
  class="flex {orientation === 'vertical'
    ? 'flex-col'
    : 'flex-row flex-wrap'} gap-1"
  use:swapListener={{
    condition: !!dragData,
    ghostIndex: _ghostIndex,
    canMixTypes: true,
    chipType: undefined,
    orientation,
  }}
  style:--ghost-width="{draggedItemWidth}px"
>
  {#each items as item, i (item)}
    {#if i === ghostIndex}
      <div
        class="ghost h-[26px] bg-gray-100 border rounded-sm pointer-events-none"
        class:!rounded-full={dragData?.type !== PivotChipType.Measure}
      ></div>
    {/if}
    <div
      class="drag-item"
      data-index={i}
      role="presentation"
      data-type={displayMap[item]?.type ?? "dimension"}
      onmousedown={(e) => handleMouseDown(e, item, i)}
    >
      {#if setMeasureFormatting && displayMap[item]?.type === "measure"}
        <MeasureFormatChip
          item={{
            id: item,
            title: displayMap[item]?.label || item,
            type: PivotChipType.Measure,
          }}
          removable
          grab
          fullWidth
          fmt={measureFormatting?.[item]}
          lowerIsBetter={lowerIsBetterMap[item] ?? false}
          onFormatChange={(fmt: PivotMeasureFormatting | null) =>
            setMeasureFormatting?.(item, fmt)}
          onRemove={() => handleRemove(item)}
        />
      {:else}
        <Chip
          removable
          grab
          fullWidth
          type={displayMap[item]?.type ?? "dimension"}
          onRemove={() => handleRemove(item)}
          label="{item} chip"
        >
          <span class="font-bold truncate" slot="body">
            {displayMap[item]?.label || item}
          </span>
        </Chip>
      {/if}
    </div>
  {/each}
  {#if dragData && ghostIndex === items.length}
    <div
      class="ghost h-[26px] bg-gray-100 border rounded-sm pointer-events-none"
      class:!rounded-full={dragData?.type !== PivotChipType.Measure}
    ></div>
  {/if}
</div>

{#if dragData}
  <PivotPortalItem
    {offset}
    width={draggedItemWidth}
    item={dragData}
    position={dragStart}
    removable
    onRelease={() => (dragData = null)}
  />
{/if}

<style lang="postcss">
  .ghost {
    width: var(--ghost-width);
  }
</style>
