<script lang="ts">
  import Checkbox from "./Checkbox.svelte";
  import MetricsTableRow from "./MetricsTableRow.svelte";
  import { ROW_HEIGHT } from "./lib";
  import { editingItem, YAMLDimension, YAMLMeasure } from "./lib";

  const headers = ["Name", "Label", "SQL expression", "Format", "Description"];
  const gutterWidth = 56;

  export let type: "measures" | "dimensions";
  export let items: Array<YAMLDimension | YAMLMeasure>;
  export let selected: Set<number>;
  export let editingIndex: number | null;
  export let searchValue: string;
  export let longest: { name: number; label: number };
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
  let wrapperRect = new DOMRectReadOnly(0, 0, 0, 0);

  let cursorDragStart = 0;
  let initiatingDragIndex = -1;
  let dragging: number[] = [];
  let dragMovement = 0;
  let insertIndex: number = -1;

  $: wrapperWidth = wrapperRect.width;
  $: expressionWidth = Math.max(180, wrapperRect.width * 0.25);

  $: filteredIndices = items
    .map((_, i) => i)
    .filter((i) => filter(items[i], searchValue));

  $: nameColumn = Math.max(180, longest.name * 8.5 + 32);
  $: labelColumn = Math.max(180, longest.label * 7 + 50);

  $: formatWidth = type === "measures" ? 140 : 0;

  $: partialWidth =
    gutterWidth + nameColumn + expressionWidth + labelColumn + formatWidth;

  $: descriptionWidth = Math.max(220, wrapperWidth - partialWidth);

  $: tableWidth = partialWidth + descriptionWidth;

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

  function filter(item: YAMLDimension | YAMLMeasure, searchValue: string) {
    return (
      item?.name?.toLowerCase().includes(searchValue.toLowerCase()) ||
      item?.label?.toLowerCase().includes(searchValue.toLowerCase()) ||
      item?.expression?.toLowerCase().includes(searchValue.toLowerCase()) ||
      (item instanceof YAMLDimension &&
        item?.column?.toLowerCase().includes(searchValue.toLowerCase()))
    );
  }
</script>

<div
  class="wrapper"
  style:max-height="{Math.max(80, ((filteredIndices?.length ?? 0) + 1) * 40) +
    1}px"
  on:scroll={(e) => {
    scroll = e.currentTarget.scrollLeft;
  }}
  bind:contentRect={wrapperRect}
>
  <table style:width="{tableWidth}px">
    <colgroup>
      <col
        style:width="{gutterWidth}px"
        style:min-width="{gutterWidth}px"
        style:max-width="{gutterWidth}px"
      />
      <col style:width="{nameColumn}px" style:min-width="{nameColumn}px" />
      <col style:width="{labelColumn}px" style:min-width="{labelColumn}px" />
      <col
        style:width="{expressionWidth}px"
        style:min-width="{expressionWidth}px"
      />

      {#if type === "measures"}
        <col style:width="{formatWidth}px" style:min-width="{formatWidth}px" />
      {/if}

      <col
        style:min-width="{descriptionWidth}px"
        style:width="{descriptionWidth}px"
      />
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
      {#each filteredIndices as index (index)}
        <MetricsTableRow
          disableDrag={filteredIndices.length !== items.length}
          sidebarOpen={!!$editingItem}
          i={index}
          item={items[index]}
          {type}
          {expressionWidth}
          selected={selected.has(index)}
          dragging={!!cursorDragStart}
          tableWidth={tableWidth - wrapperWidth}
          scrollLeft={scroll}
          length={items.length}
          editing={editingIndex === index}
          {onEdit}
          {onDuplicate}
          {onDelete}
          handleDragStart={(event) => handleDragStart(event, index)}
          onCheckedChange={(checked) => {
            onCheckedChange(checked, index);
          }}
          onMoveTo={(top) => {
            const moving = Array.from(selected);

            if (!selected.has(index)) {
              moving.push(index);
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
    @apply overflow-x-auto overflow-y-hidden relative;
    @apply border rounded-[2px];
    @apply w-full;
  }

  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-full;
    @apply font-normal select-none relative h-fit;
    @apply table-fixed;
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
    @apply truncate;
  }
  .row-insert-marker {
    @apply w-full h-[3px] bg-primary-300 absolute z-50;
    @apply -translate-y-1/2;
  }

  .last {
    @apply -translate-y-full;
  }
</style>
