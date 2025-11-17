<script context="module" lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Column from "@rilldata/web-common/components/icons/Column.svelte";
  import Row from "@rilldata/web-common/components/icons/Row.svelte";
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { writable } from "svelte/store";
  import { getStateManagers } from "../state-managers/state-managers";
  import { metricsExplorerStore } from "../stores/dashboard-stores";
  import AddField from "./AddField.svelte";
  import PivotChip from "./PivotChip.svelte";
  import PivotPortalItem from "./PivotPortalItem.svelte";
  import { swapListener } from "./swapListener";
  import TimeDropdownChip from "./TimeDropdownChip.svelte";
  import {
    type PivotChipData,
    PivotChipType,
    type PivotTableMode,
  } from "./types";

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
  export let tableMode: PivotTableMode = "nest";
  export let onUpdate: (items: PivotChipData[]) => void = () => {};

  import {
    handleTimeChipClick,
    handleTimeChipDrop,
    isNewTimeChip,
    updateTimeChipGrain,
  } from "@rilldata/web-common/features/dashboards/pivot/time-pill-utils";
  import { timePillSelectors } from "./time-pill-store";

  const isDropLocation = zone === "columns" || zone === "rows";

  const _ghostIndex = writable<number | null>(null);

  let swap = false;
  let container: HTMLDivElement;
  let offset = { x: 0, y: 0 };
  let dragStart = { left: 0, top: 0 };

  const { exploreName } = getStateManagers();

  $: ghostIndex = $_ghostIndex;
  $: dragData = $dragDataStore;
  $: source = dragData?.source;
  $: dragChip = dragData?.chip;
  $: ghostWidth = dragData?.width;
  $: initialIndex = dragData?.initialIndex ?? -1;
  $: canMixTypes = zone === "columns" && tableMode === "flat";
  $: zoneStartedDrag = source === zone;
  $: lastDimensionIndex = items.findLastIndex(
    (i) => i.type !== PivotChipType.Measure,
  );

  $: isValidDropZone =
    isDropLocation &&
    dragData &&
    (zone === "columns" || dragChip?.type !== PivotChipType.Measure);

  // Get available grains from the store
  const availableGrainsStore = timePillSelectors.getAvailableGrains("time");
  $: availableTimeGrains = $availableGrainsStore;

  function handleMouseDown(e: MouseEvent, item: PivotChipData) {
    const target = e.target as HTMLElement;
    if (target.closest(".grain-dropdown") || target.closest(".grain-label"))
      return;

    e.preventDefault();

    if (e.button !== 0) return;

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
    _ghostIndex.set(index);

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
          onUpdate(temp);
        },
        {
          once: true,
          signal: controller.signal,
        },
      );
    }

    window.addEventListener("mouseup", reset, {
      once: true,
    });

    dragDataStore.set({
      chip: item,
      source: zone,
      width,
      initialIndex,
    });
  }

  function reset() {
    dragDataStore.set(null);
    _ghostIndex.set(null);
  }

  function handleDrop() {
    if (zoneStartedDrag)
      $controllerStore?.abort("Drag cancelled - item dropped");

    if (isValidDropZone) {
      if (dragChip && ghostIndex !== null) {
        const temp = [...items];

        let chipToAdd = dragChip;

        if (isNewTimeChip(chipToAdd)) {
          const timeChipsInZone = temp.filter(
            (chip) => chip.type === PivotChipType.Time,
          );

          chipToAdd = handleTimeChipDrop(
            dragChip,
            ghostIndex,
            timeChipsInZone,
            availableTimeGrains,
          );
        }

        temp.splice(ghostIndex, 0, chipToAdd);
        items = temp;
        onUpdate(items);
      }
      swap = false;
    }
    reset();
  }

  function handleDragEnter() {
    if (!dragData) return;

    if (!isValidDropZone) return;

    const defaultIndex =
      dragChip?.type === PivotChipType.Measure
        ? items.length
        : lastDimensionIndex + 1;

    _ghostIndex.set(defaultIndex);
    swap = true;
  }

  function handleDragLeave() {
    if (!dragData) return;
    if (zone === "columns" || zone === "rows") {
      _ghostIndex.set(null);
    }

    swap = false;
  }

  function handleRowClick(item: PivotChipData) {
    let itemToAdd = item;
    if (item.type === PivotChipType.Time) {
      itemToAdd = handleTimeChipClick(item, availableTimeGrains);
    }
    metricsExplorerStore.addPivotField($exploreName, itemToAdd, true);
  }

  function handleColumnClick(item: PivotChipData) {
    let itemToAdd = item;
    if (item.type === PivotChipType.Time) {
      itemToAdd = handleTimeChipClick(item, availableTimeGrains);
    }
    metricsExplorerStore.addPivotField($exploreName, itemToAdd, false);
  }

  function handleTimeGrainSelect(item: PivotChipData, timeGrain: V1TimeGrain) {
    const updatedItems = updateTimeChipGrain(items, item, timeGrain);
    items = updatedItems;
    onUpdate(updatedItems);
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
    ghostIndex: _ghostIndex,
    chipType: dragChip?.type,
    canMixTypes,
    orientation: "horizontal",
  }}
  bind:this={container}
>
  {#each items as item, index (item.id)}
    <div
      class="item-wrapper gap-x-2"
      class:aligned={zone === "Time" ||
        zone === "Measures" ||
        zone === "Dimensions"}
    >
      {#if index === ghostIndex}
        <span
          class="ghost"
          class:rounded={dragChip?.type !== PivotChipType.Measure}
        />
      {/if}

      <div
        id={item.id}
        data-type={item.type === PivotChipType.Measure
          ? "measure"
          : "dimension"}
        data-index={index}
        class="drag-item"
        class:hidden={dragChip?.id === item.id && zoneStartedDrag}
        class:rounded-full={item.type !== PivotChipType.Measure}
      >
        {#if isDropLocation && item.type === PivotChipType.Time}
          <TimeDropdownChip
            {item}
            grab
            removable
            availableGrains={availableTimeGrains}
            onTimeGrainSelect={(timeGrain) =>
              handleTimeGrainSelect(item, timeGrain)}
            on:mousedown={(e) => handleMouseDown(e, item)}
            onRemove={() => {
              items = items.filter((i) => i.id !== item.id);
              onUpdate(items);
            }}
          />
        {:else}
          <PivotChip
            {item}
            grab
            removable={isDropLocation}
            on:mousedown={(e) => handleMouseDown(e, item)}
            onRemove={() => {
              items = items.filter((i) => i.id !== item.id);
              onUpdate(items);
            }}
          />
        {/if}
      </div>

      {#if zone !== "rows" && zone !== "columns"}
        <div class="icons">
          {#if (zone === "Time" || zone === "Dimensions") && tableMode === "nest"}
            <Tooltip distance={8} location="top" alignment="start">
              <button
                class="icon-wrapper"
                on:click={() => handleRowClick(item)}
                aria-label="Add Row"
                type="button"
              >
                <Row size="16px" />
              </button>
              <TooltipContent slot="tooltip-content">Add to rows</TooltipContent
              >
            </Tooltip>
          {/if}

          <Tooltip distance={8} location="top" alignment="start">
            <button
              class="icon-wrapper"
              on:click={() => handleColumnClick(item)}
              aria-label="Add Column"
              type="button"
            >
              <Column size="16px" />
            </button>
            <TooltipContent slot="tooltip-content">
              Add to columns
            </TooltipContent>
          </Tooltip>
        </div>
      {/if}
    </div>
  {:else}
    {#if ghostIndex === null}
      <p>{placeholder}</p>
    {/if}
  {/each}

  {#if ghostIndex === items.length}
    <span
      class="ghost"
      class:rounded={dragChip?.type !== PivotChipType.Measure}
    />
  {/if}

  {#if zone === "columns" || zone === "rows"}
    <AddField {zone} />
    {#if items.length}
      <Button
        type="text"
        onClick={() => {
          onUpdate([]);
        }}
      >
        Clear
      </Button>
    {/if}
  {/if}
</div>

{#if dragChip && zoneStartedDrag}
  <PivotPortalItem
    {offset}
    item={dragChip}
    position={dragStart}
    removable={isDropLocation}
    onRelease={() => dragDataStore.set(null)}
  />
{/if}

<style lang="postcss">
  .ghost {
    @apply bg-gray-100 border rounded-sm pointer-events-none;
    height: 26px;
    width: var(--ghost-width);
  }

  .dnd-zone {
    @apply w-full max-w-full rounded-sm;
    @apply flex flex-col;
    @apply gap-y-2 py-2  text-gray-500;
  }

  .horizontal {
    @apply flex flex-row flex-wrap bg-gray-50 w-full p-1 px-2 gap-x-2 h-fit;
    @apply items-center;
    @apply border;
  }

  .valid {
    @apply border-blue-400;
  }

  .valid:hover {
    @apply bg-surface;
  }

  .rounded {
    @apply rounded-full;
  }

  .item-wrapper {
    @apply flex items-center;
  }

  .item-wrapper.aligned {
    @apply justify-between w-full;
  }

  .icons {
    @apply flex gap-x-2 opacity-0 transition-opacity duration-200;
  }

  .item-wrapper:hover .icons {
    @apply opacity-100;
  }

  .icon-wrapper {
    @apply inline-flex items-center justify-center cursor-pointer;
  }
</style>
