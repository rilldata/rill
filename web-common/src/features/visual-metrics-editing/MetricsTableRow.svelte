<script context="module" lang="ts">
  import { writable } from "svelte/store";

  export const insertIndex = writable<number | null>(null);
  export const table = writable<"dimensions" | "measures" | null>(null);
</script>

<script lang="ts">
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import { measureChipColors as colors } from "@rilldata/web-common/components/chip/chip-types";
  import EditControls from "./EditControls.svelte";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import Checkbox from "./Checkbox.svelte";
  import { editingItem } from "../workspaces/VisualMetrics.svelte";
  import {
    defaultChipColors,
    measureChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";
  import { YAMLMap } from "yaml";

  const ROW_HEIGHT = 40;

  export let item: YAMLMap<string, string>;
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

    clone.classList.remove("row");
    clone.style.opacity = "0.6";
    // clone.style.position = "fixed";
    // clone.style.display = "table-row";
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

  // function isMeasure(
  //   item: MetricsViewSpecDimensionV2 | MetricsViewSpecMeasureV2,
  // ): item is MetricsViewSpecMeasureV2 {
  //   return "formatPreset" in item;
  // }

  function setEditing() {
    editingItem.set({ index: i, type });
  }

  $: editing = $editingItem?.index === i && $editingItem?.type === type;
</script>

<tr
  id={item.get("name") || item.get("label")}
  style:transform="translateY(0px)"
  class="relative text-sm row"
  style:height="{ROW_HEIGHT}px"
  class:editing
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
        class="text-gray-500"
      >
        <DragHandle size="16px" />
      </button>
      <Checkbox onChange={onCheckedChange} checked={selected} />
    </div>
  </td>
  <td class="max-w-64 source-code" on:click={setEditing}>
    {item?.get("name")}</td
  >

  <td class="source-code max-w-72" on:click={setEditing}>
    {item.get("expression") || item.get("column")}
  </td>
  <td on:click={setEditing}>
    <div class="pointer-events-none text-[12px]">
      <Chip
        {...colors}
        slideDuration={0}
        extraRounded={false}
        extraPadding={false}
        {...type === "measures" ? measureChipColors : defaultChipColors}
        label={item.get("label") || item.get("name")}
        outline
      >
        <span slot="body" class="font-bold truncate">
          {item.get("label") || item.get("name")}
        </span>
      </Chip>
    </div>
  </td>
  {#if type === "measures"}
    <td on:click={setEditing}>
      {item.get("format_preset") || item?.get("format_d3") || "-"}</td
    >
  {/if}
  <td class="max-w-72" on:click={setEditing}>
    {item?.get("description") || "-"}
  </td>

  {#if hovered}
    <EditControls
      {editing}
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

  tr:hover:not(.editing) {
    @apply bg-gray-50;
  }

  .editing {
    @apply bg-gray-100;
  }

  td:not(.dragging) {
    @apply pl-4 truncate border-b;
  }

  .insert td {
    @apply border-b border-primary-500;
  }

  .row:last-of-type td {
    @apply border-b-0;
  }

  .source-code {
    font-family: "Source Code Variable", monospace;
  }
</style>
