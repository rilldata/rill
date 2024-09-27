<script lang="ts">
  import { onMount } from "svelte";
  import Checkbox from "./Checkbox.svelte";
  import MetricsTableRow from "./MetricsTableRow.svelte";
  import { nameWidth, labelWidth, formatWidth, ROW_HEIGHT } from "./lib";
  import { editingItem, YAMLDimension, YAMLMeasure } from "./lib";

  const headers = ["Name", "Label", "SQL expression", "Format", "Description"];
  const gutterWidth = 56;

  export let type: "measures" | "dimensions";
  export let items: Array<YAMLDimension | YAMLMeasure>;
  export let selected: Set<number>;
  export let editingIndex: number | null;
  export let reorderList: (
    initIndex: number[],
    newIndex: number,
    type: "measures" | "dimensions",
  ) => void;
  export let onDuplicate: (
    index: number,
    type: "measures" | "dimensions",
  ) => void;
  export let onDelete: (index: number, type: "measures" | "dimensions") => void;
  export let onCheckedChange: (checked: boolean, index?: number) => void;
  export let onEdit: (
    index: number,
    type: "measures" | "dimensions",
    field?: string,
  ) => void;

  let clientWidth: HTMLTableRowElement;
  let tbody: HTMLTableSectionElement;
  let scroll = 0;
  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  let wrapperRect = new DOMRectReadOnly(0, 0, 0, 0);

  let cursorDragStart = 0;
  let initiatingDragIndex = -1;
  let dragging: number[] = [];
  let dragMovement = 0;
  let insertIndex: number = -1;

  onMount(() => {
    const cells = clientWidth.children;
    const initialNameWidth = cells[1].getBoundingClientRect().width;
    const initialLabelWidth = cells[2].getBoundingClientRect().width;
    const initialFormatWidth =
      type === "measures" ? cells[4].getBoundingClientRect().width : 0;

    nameWidth.set(initialNameWidth);
    labelWidth.set(initialLabelWidth);
    formatWidth.set(initialFormatWidth);
  });

  $: tableWidth = contentRect.width;
  $: wrapperWidth = wrapperRect.width;
  $: expressionWidth = Math.max(220, wrapperRect.width * 0.2);

  function handleDragStart(e: MouseEvent, i: number) {
    if (e.button !== 0) return;

    cursorDragStart = e.clientY;
    initiatingDragIndex = i;
    insertIndex = initiatingDragIndex + 1;

    dragging = Array.from(selected);

    if (!selected.has(i)) {
      dragging.push(i);
    }

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);
  }

  function handleMouseMove(e: MouseEvent) {
    dragMovement = e.clientY - cursorDragStart;
    const rowDelta = Math.floor(dragMovement / 40);

    insertIndex = Math.max(0, initiatingDragIndex + 1 + rowDelta);
  }

  function handleMouseUp() {
    window.removeEventListener("mousemove", handleMouseMove);
    window.removeEventListener("mouseup", handleMouseUp);

    if (insertIndex !== initiatingDragIndex && insertIndex !== null) {
      reorderList(dragging, insertIndex, type);
    }

    resetDrag();
  }

  function resetDrag() {
    cursorDragStart = 0;
    initiatingDragIndex = -1;
    dragging = [];
    dragMovement = 0;
    insertIndex = -1;
  }
</script>

<div
  class="wrapper relative"
  style:max-height="{Math.max(80, ((items?.length ?? 0) + 1) * 40) + 1}px"
  on:scroll={(e) => {
    scroll = e.currentTarget.scrollLeft;
  }}
  bind:contentRect={wrapperRect}
>
  <table bind:contentRect>
    <colgroup>
      <col style:width="{gutterWidth}px" style:min-width="{gutterWidth}px" />
      <col style:width="{$nameWidth}px" style:min-width="{$nameWidth}px" />
      <col style:width="{$labelWidth}px" style:min-width="{$labelWidth}px" />
      <col
        style:width="{expressionWidth}px"
        style:min-width="{expressionWidth}px"
      />

      {#if type === "measures"}
        <col
          style:width="{$formatWidth}px"
          style:min-width="{$formatWidth}px"
        />
      {/if}

      <col />
    </colgroup>

    <thead class="sticky top-0 z-10">
      <tr bind:this={clientWidth}>
        <th class="!pl-[22px]">
          <Checkbox
            onChange={onCheckedChange}
            checked={selected.size / items.length}
          />
        </th>
        {#each headers as header (header)}
          {#if (type === "dimensions" && header !== "Format") || type === "measures"}
            <th>
              {header}
            </th>
          {/if}
        {/each}
      </tr>
    </thead>
    <tbody bind:this={tbody} class="relative overflow-hidden">
      {#each items as item, i (i)}
        <MetricsTableRow
          sidebarOpen={!!$editingItem}
          {i}
          {item}
          {type}
          {expressionWidth}
          selected={selected.has(i)}
          dragging={!!cursorDragStart}
          tableWidth={tableWidth - wrapperWidth}
          scrollLeft={scroll}
          length={items.length}
          editing={editingIndex === i}
          {onEdit}
          {onDuplicate}
          {onDelete}
          handleDragStart={(event) => handleDragStart(event, i)}
          onCheckedChange={(checked) => {
            onCheckedChange(checked, i);
          }}
          onMoveTo={(top) => {
            const moving = Array.from(selected);

            if (!selected.has(i)) {
              moving.push(i);
            }
            reorderList(moving, top ? 0 : items.length, type);
          }}
        />
      {:else}
        <tr style:height="40px" class="relative">
          <div
            class="absolute left-0 h-10 px-6 items-center flex w-full italic"
          >
            No items matching search
          </div>
        </tr>
      {/each}

      {#each dragging as i, dragIndex (i)}
        {@const item = items[i]}
        <MetricsTableRow
          ghost
          {item}
          {type}
          {i}
          translate={dragMovement + (items.length - i + dragIndex) * -40}
          {expressionWidth}
          tableWidth={tableWidth - wrapperWidth}
          selected={selected.has(i)}
        />
      {/each}
    </tbody>
  </table>
  {#if insertIndex !== -1}
    {@const last = insertIndex === items.length}
    <span
      style:top="{insertIndex * ROW_HEIGHT + ROW_HEIGHT - (last ? 1 : 0)}px"
      class:last
      class="row-insert-marker"
    />
  {/if}
</div>

<style lang="postcss">
  .wrapper {
    @apply overflow-x-auto overflow-y-hidden w-full max-w-full relative;
    @apply border rounded-[2px] min-h-fit h-fit;
  }

  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-full;
    @apply font-normal select-none relative h-fit;
  }

  tbody {
    @apply cursor-pointer;
  }

  thead tr {
    height: 40px !important;
  }

  th {
    @apply text-left;
    @apply pl-4 text-slate-500 bg-background;
    @apply border-b text-sm font-semibold;
  }
  .row-insert-marker {
    @apply w-full h-[3px] bg-primary-300 absolute z-50;
    @apply -translate-y-1/2;
  }

  .last {
    @apply -translate-y-full;
  }
</style>
