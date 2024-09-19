<script context="module" lang="ts">
  class MaxStore {
    private store: Writable<number> = writable(0);

    set(value: number) {
      this.store.set(Math.max(value, get(this.store)));
    }

    subscribe = this.store.subscribe;
  }
  const nameWidth = new MaxStore();
  const labelWidth = new MaxStore();
  const formatWidth = new MaxStore();
</script>

<script lang="ts">
  import { onMount } from "svelte";
  import Checkbox from "./Checkbox.svelte";
  import MetricsTableRow from "./MetricsTableRow.svelte";
  import { insertIndex, table } from "./MetricsTableRow.svelte";
  import { YAMLMap } from "yaml";
  import { get, Writable, writable } from "svelte/store";

  const headers = ["Name", "Label", "SQL expression", "Format", "Description"];
  const gutterWidth = 56;
  const ROW_HEIGHT = 40;

  export let dimensions: boolean = false;
  export let items: Array<YAMLMap<string, string>>;
  export let reorderList: (
    initIndex: number,
    newIndex: number,
    type: "measures" | "dimensions",
  ) => void;
  export let onDuplicate: (
    index: number,
    type: "measures" | "dimensions",
  ) => void;
  export let onDelete: (index: number, type: "measures" | "dimensions") => void;

  onMount(() => {
    const cells = clientWidth.children;
    const initialNameWidth = cells[1].getBoundingClientRect().width;
    const initialLabelWidth = cells[2].getBoundingClientRect().width;
    const initialFormatWidth = !dimensions
      ? cells[4].getBoundingClientRect().width
      : 0;

    nameWidth.set(initialNameWidth);
    labelWidth.set(initialLabelWidth);
    formatWidth.set(initialFormatWidth);
  });

  let clientWidth: HTMLTableRowElement;

  let tbody: HTMLTableSectionElement;
  let selected = new Set();
  let scroll = 0;
  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  let wrapperRect = new DOMRectReadOnly(0, 0, 0, 0);

  $: type = (dimensions ? "dimensions" : "measures") as
    | "dimensions"
    | "measures";

  $: tableWidth = contentRect.width;
  $: wrapperWidth = wrapperRect.width;
  $: expressionWidth = Math.max(220, wrapperRect.width * 0.2);
</script>

<div
  class="wrapper relative"
  on:scroll={(e) => {
    scroll = e.currentTarget.scrollLeft;
  }}
  bind:contentRect={wrapperRect}
  style:max-height="{Math.max(80, ((items?.length ?? 0) + 1) * 40)}px"
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

      {#if !dimensions}
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
            onChange={(checked) => {
              if (checked) {
                selected = new Set(
                  Array.from({ length: items.length }, (_, i) => i),
                );
              } else {
                selected = new Set();
              }
            }}
          />
        </th>
        {#each headers as header (header)}
          {#if (dimensions && header !== "Format") || !dimensions}
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
          rowHeight={ROW_HEIGHT}
          {expressionWidth}
          {item}
          {reorderList}
          {onDuplicate}
          tableWidth={tableWidth - wrapperWidth}
          {i}
          scrollLeft={scroll}
          length={items.length}
          onCheckedChange={(checked) => {
            selected[checked ? "add" : "delete"](i);
          }}
          selected={selected.has(i)}
          {type}
          {onDelete}
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
    </tbody>
  </table>
  {#if $insertIndex !== null && $table === type}
    <span
      style:top="{($insertIndex + 1) * ROW_HEIGHT + ROW_HEIGHT}px"
      class="w-full h-[3px] bg-primary-300 absolute top-[40px] -translate-y-1/2 z-50"
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
</style>
