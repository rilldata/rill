<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import PivotPortalItem from "@rilldata/web-common/features/dashboards/pivot/PivotPortalItem.svelte";
  import { swapListener } from "@rilldata/web-common/features/dashboards/pivot/swapListener";
  import {
    PivotChipType,
    type PivotChipData,
  } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { writable } from "svelte/store";
  import type { FieldType } from "./types";

  const _ghostIndex = writable<number | null>(null);

  export let items: string[] = [];
  export let displayMap: Record<string, { label: string; type: FieldType }>;
  export let onUpdate: (items: string[]) => void;
  export let orientation: "horizontal" | "vertical" = "vertical";

  let dragData: PivotChipData | null = null;
  let dragStart = { left: 0, top: 0 };
  let offset = { x: 0, y: 0 };
  let draggedItemWidth = 0;

  $: ghostIndex = $_ghostIndex;

  function handleMouseDown(e: MouseEvent, item: string, index: number) {
    // No-op if the target is a chip remove button
    const target = e.target as HTMLElement;
    if (target.closest('[aria-label="Remove"]')) return;

    e.preventDefault();
    const dragItem = e.currentTarget as HTMLElement;
    if (!dragItem) return;

    const { width, left, top } = dragItem.getBoundingClientRect();
    dragStart = { left, top };
    offset = {
      x: e.clientX - left,
      y: e.clientY - top,
    };

    dragData = {
      id: item,
      title: displayMap[item]?.label?.replace(/Time /, "") || item,
      type:
        displayMap[item]?.type === "measure"
          ? PivotChipType.Measure
          : displayMap[item]?.type === "time"
            ? PivotChipType.Time
            : PivotChipType.Dimension,
    };
    draggedItemWidth = width;
    _ghostIndex.set(index);

    const temp = [...items];
    temp.splice(index, 1);
    items = temp;

    window.addEventListener(
      "mouseup",
      () => {
        handleDrop();
      },
      { once: true },
    );
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
      />
    {/if}
    <div
      class="drag-item"
      data-index={i}
      role="presentation"
      data-type={displayMap[item]?.type ?? "dimension"}
      on:mousedown={(e) => handleMouseDown(e, item, i)}
    >
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
    </div>
  {/each}
  {#if dragData && ghostIndex === items.length}
    <div
      class="ghost h-[26px] bg-gray-100 border rounded-sm pointer-events-none"
      class:!rounded-full={dragData?.type !== PivotChipType.Measure}
    />
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
