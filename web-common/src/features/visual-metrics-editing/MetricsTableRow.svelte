<script context="module" lang="ts">
  import { writable } from "svelte/store";

  export const insertIndex = writable<number | null>(null);
  export const table = writable<"dimensions" | "measures" | null>(null);
</script>

<script lang="ts">
  import type {
    MetricsViewSpecDimensionV2,
    MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import { measureChipColors as colors } from "@rilldata/web-common/components/chip/chip-types";
  import EditControls from "./EditControls.svelte";
  import { GripVertical } from "lucide-svelte";
  import Checkbox from "./Checkbox.svelte";
  import { editingItem } from "../workspaces/VisualMetrics.svelte";
  import {
    defaultChipColors,
    measureChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";

  const ROW_HEIGHT = 40;

  export let item: MetricsViewSpecMeasureV2 | MetricsViewSpecDimensionV2;
  export let i: number;
  export let selected: boolean;
  export let length: number;
  export let reorderList: (
    initIndex: number,
    newIndex: number,
    type: "measures" | "dimensions",
  ) => void;
  export let onCheckedChange: (checked: boolean) => void;
  export let onDelete: (index: number, type: "measures" | "dimensions") => void;
  export let onDuplicate: (
    index: number,
    type: "measures" | "dimensions",
  ) => void;
  export let scrollLeft: number;
  export let tableWidth: number;
  export let type: "measures" | "dimensions";

  let row: HTMLTableRowElement;
  let initialY = 0;
  let clone: HTMLTableRowElement;

  function handleDragStart(e: MouseEvent) {
    if (e.button !== 0) return;
    table.set(type);
    initialY = e.clientY;

    clone = row.cloneNode(true) as HTMLTableRowElement;

    clone.style.opacity = "0.6";
    clone.style.position = "fixed";
    clone.style.display = "table-row";
    clone.style.width = "100%";
    clone.style.transform = `translateY(${e.clientY - initialY - (length - i) * 40}px)`;
    row.parentElement?.appendChild(clone);

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);
  }

  function handleMouseMove(e: MouseEvent) {
    const movement = e.clientY - initialY;
    const rowDelta = Math.floor(movement / 40);

    insertIndex.set(i + rowDelta);
    clone.style.transform = `translateY(${e.clientY - initialY - (length - i) * 40}px)`;
  }

  function handleMouseUp() {
    window.removeEventListener("mousemove", handleMouseMove);
    window.removeEventListener("mouseup", handleMouseUp);

    if ($insertIndex !== i && $insertIndex !== null) {
      reorderList(i, $insertIndex < i ? $insertIndex + 1 : $insertIndex, type);
    }

    clone.remove();
    table.set(null);
    insertIndex.set(null);
  }
  let hovered = false;

  function isMeasure(
    item: MetricsViewSpecDimensionV2 | MetricsViewSpecMeasureV2,
  ): item is MetricsViewSpecMeasureV2 {
    return "formatPreset" in item;
  }

  function setEditing() {
    editingItem.set({ data: item, index: i, type });
  }
</script>

<tr
  id={item?.name || item?.label}
  style:transform="translateY(0px)"
  class="relative text-sm"
  style:height="{ROW_HEIGHT}px"
  class:editing={$editingItem?.index === i && $editingItem?.type === type}
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
  bind:this={row}
  class:insert={$table === type && $insertIndex === i}
>
  <td class="!pl-0">
    <div class="gap-x-0.5 flex items-center w-14 pl-1">
      <button
        on:mousedown={handleDragStart}
        class:opacity-0={!hovered}
        disabled={!hovered}
      >
        <GripVertical size="14px" class="text-gray-500" />
      </button>
      <Checkbox onChange={onCheckedChange} checked={selected} />
    </div>
  </td>
  <td class="max-w-64 source-code" on:click={setEditing}>
    {item?.name || item?.label}</td
  >

  <td class="source-code max-w-72" on:click={setEditing}
    >{item?.expression || item?.name}</td
  >
  <td on:click={setEditing}>
    <div class="pointer-events-none text-[12px]">
      <Chip
        {...colors}
        slideDuration={0}
        extraRounded={false}
        extraPadding={false}
        {...isMeasure(item) ? measureChipColors : defaultChipColors}
        label={item?.label || item?.label}
        outline
      >
        <span slot="body" class="font-bold truncate">
          {item?.label || item?.label}
        </span>
      </Chip>
    </div>
  </td>
  {#if isMeasure(item)}
    <td class="capitalize" on:click={setEditing}>
      {item?.formatPreset || item?.formatD3 || "-"}</td
    >
  {/if}
  <td class="max-w-72" on:click={setEditing}
    >{item?.description || item?.description || "-"}</td
  >

  {#if hovered}
    <EditControls
      right={Math.max(0, tableWidth - scrollLeft)}
      first={i === 0}
      last={i === length - 1}
      onEdit={setEditing}
      onMoveToBottom={() => {
        reorderList(i, length - 1, type);
      }}
      onMoveToTop={() => {
        reorderList(i, 0, type);
      }}
      onDuplicate={() => {
        onDuplicate(i, type);
      }}
      onDelete={() => {
        onDelete(i, type);
      }}
    />
  {/if}
</tr>

<style lang="postcss">
  tr {
    @apply bg-background;
    /* @apply -z-10; */
  }

  tr:hover,
  .editing {
    @apply bg-gray-50;
  }

  td:not(.dragging) {
    @apply pl-4 truncate border-b;
  }

  .insert td {
    @apply border-b border-primary-500;
  }

  tr:last-of-type td {
    @apply border-b-0;
  }

  .source-code {
    font-family: "Source Code Variable", monospace;
  }
</style>
