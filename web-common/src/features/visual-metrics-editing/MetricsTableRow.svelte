<script context="module" lang="ts">
  import { writable } from "svelte/store";

  export const insertIndex = writable<number | null>(null);
  export const table = writable<"dimensions" | "measures" | null>(null);
</script>

<script lang="ts">
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import EditControls from "./EditControls.svelte";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import Checkbox from "./Checkbox.svelte";
  import { editingItem } from "../workspaces/VisualMetrics.svelte";
  import {
    defaultChipColors,
    measureChipColors,
  } from "@rilldata/web-common/components/chip/chip-types";
  import { YAMLMap } from "yaml";

  export let rowHeight: number;
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
  export let expressionWidth: number;

  let row: HTMLTableRowElement;
  let initialY = 0;
  let clone: HTMLTableRowElement;

  function handleDragStart(e: MouseEvent) {
    if (e.button !== 0) return;
    table.set(type);
    initialY = e.clientY;

    clone = row.cloneNode(true) as HTMLTableRowElement;

    clone.style.opacity = "0.6";
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

  function setEditing(
    e: MouseEvent & {
      currentTarget: EventTarget & HTMLTableCellElement;
    },
  ) {
    const field = e.currentTarget?.getAttribute("aria-label");
    editingItem.set({ index: i, type, field });
  }

  $: editing = $editingItem?.index === i && $editingItem?.type === type;
</script>

<tr
  id={item.get("name") || item.get("label")}
  style:transform="translateY(0px)"
  class="relative text-sm"
  style:height="{rowHeight}px"
  class:editing
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
  bind:this={row}
  class:insert={$table === type && $insertIndex === i}
>
  <td class="!pl-0">
    <div class="gap-x-0.5 flex items-center pl-1">
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

  <td class="source-code truncate" on:click={setEditing} aria-label="Name">
    <span>{item?.get("name") ?? "-"}</span>
  </td>
  <td on:click={setEditing} aria-label="Label">
    <div class="pointer-events-none text-[12px] pr-4">
      <Chip
        slideDuration={0}
        extraRounded={type === "dimensions"}
        extraPadding={false}
        {...type === "measures" ? measureChipColors : defaultChipColors}
        label={item.get("label") || item.get("name")}
        outline
      >
        <div slot="body" class="font-bold">
          {item.get("label") || item.get("name")}
        </div>
      </Chip>
    </div>
  </td>

  <td
    class="source-code truncate"
    on:click={setEditing}
    aria-label="SQL Expression"
    style:max-width="{expressionWidth}px"
  >
    <span>{item.get("expression") || item.get("column")}</span>
  </td>

  {#if type === "measures"}
    <td on:click={setEditing} aria-label="Format">
      <span>{item.get("format_preset") || item?.get("format_d3") || "-"}</span>
    </td>
  {:else}
    <!-- <td /> -->
  {/if}

  <td class="" on:click={setEditing} aria-label="Description">
    <span>{item?.get("description") || "-"}</span>
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

  span {
    @apply pr-4;
  }
</style>
