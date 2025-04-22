<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import PivotPortalItem from "@rilldata/web-common/features/dashboards/pivot/PivotPortalItem.svelte";
  import { swapListener } from "@rilldata/web-common/features/dashboards/pivot/swapListener";
  import { PivotChipType } from "@rilldata/web-common/features/dashboards/pivot/types";
  import { writable } from "svelte/store";
  import type { FieldType } from "./types";

  export let items: string[] = [];
  export let displayMap: Record<string, { label: string; type: FieldType }>;
  export let onUpdate: (items: string[]) => void;

  const ghostIndex = writable<number | null>(null);
  let draggedItem: string | null = null;
  let isDragging = false;
  let dragStart = { left: 0, top: 0 };
  let offset = { x: 0, y: 0 };
  let draggedItemWidth = 0;

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

    isDragging = true;
    draggedItem = item;
    draggedItemWidth = width;
    ghostIndex.set(index);

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
    if (draggedItem && $ghostIndex !== null) {
      if (!items.includes(draggedItem)) {
        const temp = [...items];
        temp.splice($ghostIndex, 0, draggedItem);
        items = temp;
        onUpdate(items);
      } else {
        console.warn(
          "Prevented duplicate addition of item:",
          displayMap[draggedItem]?.label || draggedItem,
        );
      }
    }
    isDragging = false;
    draggedItem = null;
    ghostIndex.set(null);
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
  class="flex flex-col gap-1"
  use:swapListener={{
    condition: isDragging,
    ghostIndex,
    canMixTypes: true,
    chipType: undefined,
  }}
  style:--ghost-width="{draggedItemWidth}px"
>
  {#each items as item, i (item)}
    {#if i === $ghostIndex}
      <div class="ghost h-[26px] bg-gray-200 rounded-sm pointer-events-none" />
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
        on:remove={() => handleRemove(item)}
      >
        <span class="font-bold truncate" slot="body">
          {displayMap[item]?.label || item}
        </span>
      </Chip>
    </div>
  {/each}
  {#if isDragging && $ghostIndex === items.length}
    <div class="ghost h-[26px] bg-gray-200 rounded-sm pointer-events-none" />
  {/if}
</div>

{#if draggedItem && isDragging}
  <PivotPortalItem
    {offset}
    width={draggedItemWidth}
    item={{
      id: draggedItem,
      title: displayMap[draggedItem]?.label || draggedItem,
      type:
        displayMap[draggedItem]?.type === "measure"
          ? PivotChipType.Measure
          : displayMap[draggedItem]?.type === "time"
            ? PivotChipType.Time
            : PivotChipType.Dimension,
    }}
    position={dragStart}
    removable
    on:release={() => (draggedItem = null)}
  />
{/if}

<style lang="postcss">
  .ghost {
    width: var(--ghost-width);
  }
</style>
