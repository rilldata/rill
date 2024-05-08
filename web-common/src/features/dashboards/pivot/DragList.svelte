<script context="module" lang="ts">
  import AddField from "./AddField.svelte";
  import PivotDragItem from "./PivotDragItem.svelte";
  import PivotPortalItem from "./PivotPortalItem.svelte";
  import { createEventDispatcher } from "svelte";
  import { PivotChipData, PivotChipType } from "./types";
  import { writable } from "svelte/store";
  import { swapListener } from "./swapListener";

  export type Zone = "rows" | "columns" | "Time" | "Measures" | "Dimensions";

  export type DragData = {
    source: Zone;
    width: number;
    chip: PivotChipData;
    initialIndex: number;
  };

  export const dragDataStore = writable<null | DragData>(null);
  export const controllerStore = writable<AbortController | null>(null);
</script>

<script lang="ts">
  export let items: PivotChipData[] = [];
  export let placeholder: string | null = null;
  export let zone: Zone;

  const isDropLocation = zone === "columns" || zone === "rows";

  const dispatch = createEventDispatcher();
  const ghostIndex = writable<number | null>(null);

  let swap = false;
  let container: HTMLDivElement;
  let offset = { x: 0, y: 0 };
  let dragStart = { left: 0, top: 0 };

  $: dragData = $dragDataStore;
  $: source = dragData?.source;
  $: dragChip = dragData?.chip;
  $: ghostWidth = dragData?.width;
  $: initialIndex = dragData?.initialIndex ?? -1;
  $: zoneStartedDrag = source === zone;
  $: lastDimensionIndex = items.findLastIndex(
    (i) => i.type !== PivotChipType.Measure,
  );

  $: isValidDropZone =
    isDropLocation &&
    dragData &&
    (zone === "columns" || dragChip?.type !== PivotChipType.Measure);

  function handleMouseDown(e: MouseEvent, item: PivotChipData) {
    e.preventDefault();

    const dragItem = document.getElementById(item.id);
    if (!dragItem) return;

    const { width, left, top } = dragItem.getBoundingClientRect();

    dragStart = { left, top };

    offset = {
      x: e.clientX - left,
      y: e.clientY - top,
    };

    const index = Number(dragItem.dataset.index);
    initialIndex = index;
    ghostIndex.set(index);

    if (isDropLocation) {
      swap = true;

      const temp = [...items];
      temp.splice(index, 1);
      items = temp;

      // Allow us to abort this update if the pill is dropped to the same location
      // This shouldn't be necessary after state management is updated
      const controller = new AbortController();

      controllerStore.set(controller);

      window.addEventListener(
        "mouseup",
        () => {
          dispatch("update", temp);
        },
        {
          once: true,
          signal: controller.signal,
        },
      );
    }

    dragDataStore.set({
      chip: item,
      source: zone,
      width,
      initialIndex,
    });
  }

  function handleDrop() {
    if (zoneStartedDrag) $controllerStore?.abort();

    if (isValidDropZone) {
      if (dragChip && $ghostIndex !== null) {
        const temp = [...items];

        temp.splice($ghostIndex, 0, dragChip);

        items = temp;

        dispatch("update", items);
      }
      swap = false;
    }
    dragDataStore.set(null);
    ghostIndex.set(null);
  }

  function handleDragEnter() {
    if (!dragData) return;

    if (zoneStartedDrag && !isDropLocation) {
      ghostIndex.set(initialIndex);
      return;
    }

    if (!isValidDropZone) return;

    const defaultIndex =
      dragChip?.type === PivotChipType.Measure
        ? items.length
        : lastDimensionIndex + 1;

    ghostIndex.set(defaultIndex);

    swap = true;
  }

  function handleDragLeave() {
    if (!dragData) return;
    $ghostIndex = null;
    swap = false;
  }
</script>

<div
  role="presentation"
  class="dnd-zone group"
  class:valid={isValidDropZone}
  class:horizontal={isDropLocation}
  style:--ghost-width="{ghostWidth ?? 0}px"
  on:mouseup={handleDrop}
  on:mouseenter={handleDragEnter}
  on:mouseleave={handleDragLeave}
  use:swapListener={{
    condition: isDropLocation && swap,
    ghostIndex,
    chipType: dragChip?.type,
  }}
  bind:this={container}
>
  {#each items as item, index (item.id)}
    {#if index === $ghostIndex}
      <span
        class="ghost"
        class:rounded={dragChip?.type !== PivotChipType.Measure}
      />
    {/if}

    <PivotDragItem
      {item}
      {index}
      removable={isDropLocation}
      hidden={dragChip?.id === item.id && zoneStartedDrag}
      on:mousedown={(e) => handleMouseDown(e, item)}
      on:remove={() => {
        items = items.filter((i) => i.id !== item.id);
        dispatch("update", items);
      }}
    />
  {:else}
    {#if $ghostIndex === null}
      <p>{placeholder}</p>
    {/if}
  {/each}

  {#if $ghostIndex === items.length}
    <span
      class="ghost"
      class:rounded={dragChip?.type !== PivotChipType.Measure}
    />
  {/if}

  {#if zone === "columns" || zone === "rows"}
    <AddField {zone} />
  {/if}
</div>

{#if dragChip && zoneStartedDrag}
  <PivotPortalItem
    {offset}
    item={dragChip}
    position={dragStart}
    removable={isDropLocation}
  />
{/if}

<style lang="postcss">
  .ghost {
    @apply bg-gray-200 rounded-sm pointer-events-none;
    height: 26px;
    width: var(--ghost-width);
  }

  .dnd-zone {
    @apply w-full max-w-full rounded-sm;
    @apply flex flex-col;
    @apply gap-y-2 py-2  text-gray-500;
  }

  .horizontal {
    @apply flex flex-row flex-wrap bg-slate-50 w-full p-1 px-2 gap-x-2 h-fit;
    @apply items-center;
    @apply border border-slate-50;
  }

  .valid {
    @apply border-blue-400;
  }

  .valid:hover {
    @apply bg-white;
  }

  .rounded {
    @apply rounded-full;
  }
</style>
