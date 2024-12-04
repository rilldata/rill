<script lang="ts">
  import Chip from "@rilldata/web-common/components/chip/core/Chip.svelte";
  import EditControls from "./EditControls.svelte";
  import DragHandle from "@rilldata/web-common/components/icons/DragHandle.svelte";
  import Checkbox from "./Checkbox.svelte";
  import { YAMLDimension, YAMLMeasure, ROW_HEIGHT } from "./lib";

  export let item: YAMLDimension | YAMLMeasure;
  export let i: number;
  export let selected: boolean;
  export let length = Infinity;
  export let tableWidth: number;
  export let type: "measures" | "dimensions";
  export let editing = false;
  export let expressionWidth: number;
  export let scrollLeft = 0;
  export let rowHeight = ROW_HEIGHT;
  export let dragging = true;
  export let translate = 0;
  export let ghost = false;
  export let sidebarOpen = false;
  export let disableDrag = false;
  export let handleDragStart: (e: MouseEvent) => void = () => {};
  export let onEdit: (
    index: number,
    type: "measures" | "dimensions",
    field?: string,
  ) => void = () => {};
  export let onCheckedChange: (checked: boolean) => void = () => {};
  export let onDelete: (
    index: number,
    type: "measures" | "dimensions",
  ) => void = () => {};
  export let onDuplicate: (
    index: number,
    type: "measures" | "dimensions",
  ) => void = () => {};
  export let onMoveTo: (top: boolean) => void = () => {};

  let row: HTMLTableRowElement;
  let hovered = false;

  $: ({ name, display_name, expression, description } = item);

  $: id = name ?? display_name ?? "";

  $: finalSelected = selected && !sidebarOpen;

  function onCellClick(
    e: MouseEvent & {
      currentTarget: EventTarget & HTMLTableCellElement;
    },
  ) {
    const field = e?.currentTarget?.getAttribute("aria-label") ?? undefined;
    onEdit(i, type, field);
  }
</script>

<tr
  {id}
  style:transform="translateY({translate}px)"
  class="relative text-sm"
  style:height="{rowHeight}px"
  class:editing
  class:dragging
  class:ghost
  class:selected={finalSelected}
  on:mouseenter={() => (hovered = true)}
  on:mouseleave={() => (hovered = false)}
  bind:this={row}
>
  <td class="!pl-0 sticky">
    <div class="gap-x-0.5 flex items-center pl-1">
      <button
        class:opacity-0={!hovered}
        disabled={!hovered || disableDrag}
        class="text-gray-500 disabled:cursor-not-allowed"
        on:mousedown={handleDragStart}
      >
        <DragHandle size="16px" />
      </button>

      <Checkbox onChange={onCheckedChange} checked={selected} />
    </div>
  </td>

  <td class="source-code truncate" on:click={onCellClick} aria-label="Name">
    {#if !name && item instanceof YAMLDimension && item.resourceName}
      <span class="text-gray-500" title="This name was inherited automatically">
        {item.resourceName}
      </span>
    {:else}
      <span>{name || "-"}</span>
    {/if}
  </td>
  <td on:click={onCellClick} aria-label="Display name">
    <div class="text-[12px] pr-4">
      <Chip
        slideDuration={0}
        type={type === "dimensions" ? "dimension" : "measure"}
        label={display_name || name}
      >
        <div slot="body" class="font-bold">
          {display_name || name}
        </div>
      </Chip>
    </div>
  </td>

  <td
    class="source-code truncate"
    on:click={onCellClick}
    aria-label="SQL expression"
    style:max-width="{expressionWidth}px"
  >
    <span>{expression || (item instanceof YAMLDimension && item?.column)}</span>
  </td>

  {#if item instanceof YAMLMeasure}
    <td on:click={onCellClick} aria-label="Format">
      <span>{item?.format_preset || item?.format_d3 || "-"}</span>
    </td>
  {/if}

  <td
    style:max-width="{expressionWidth}px"
    on:click={onCellClick}
    aria-label="Description"
  >
    <span>{description || "-"}</span>
  </td>

  {#if hovered && !ghost}
    <div
      class="editing-controls"
      style:right="{Math.max(0, tableWidth - scrollLeft)}px"
      class:editing
      class:selected={finalSelected}
    >
      <EditControls
        selected={finalSelected}
        first={i === 0}
        last={i === length - 1}
        onEdit={() => {
          onEdit(i, type);
        }}
        {onMoveTo}
        onDuplicate={() => {
          onDuplicate(i, type);
        }}
        onDelete={() => {
          onDelete(i, type);
        }}
      />
    </div>
  {/if}
</tr>

<style lang="postcss">
  .editing-controls {
    height: 39px;
    width: 192px;
    @apply bg-gray-50;
    @apply gap-x-2.5 px-4 py-2 flex items-center justify-center absolute top-0  z-50;
  }

  .editing-controls.selected {
    @apply bg-primary-50;
  }

  tr {
    @apply bg-background;
  }

  .dragging {
    @apply pointer-events-none;
  }

  tr:hover:not(.editing) {
    @apply bg-gray-50;
  }

  tr:hover.selected {
    @apply bg-primary-50;
  }

  td:hover:not(.editing) {
    @apply text-primary-700;
  }

  .editing {
    @apply bg-gray-100;
  }

  .selected {
    @apply bg-primary-50/50;
  }

  td:not(.dragging) {
    @apply pl-4 truncate border-b;
  }

  .source-code {
    font-family: "Source Code Variable", monospace;
  }

  span {
    @apply pr-4;
  }

  .ghost {
    @apply opacity-50 pointer-events-none;
  }
</style>
