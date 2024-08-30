<script lang="ts">
  import {
    MetricsViewSpecDimensionV2,
    MetricsViewSpecMeasureV2,
  } from "@rilldata/web-common/runtime-client";
  import Checkbox from "./Checkbox.svelte";
  import MetricsTableRow from "./MetricsTableRow.svelte";
  import { insertIndex } from "./MetricsTableRow.svelte";

  const headers = ["Name", "Definition", "Label", "Format", "Description"];
  const gutterWidth = 56;

  export let reorderList: (initIndex: number, newIndex: number) => void;
  export let dimensions: boolean = false;
  export let onDuplicate: (index: number) => void;
  export let items: MetricsViewSpecDimensionV2[] | MetricsViewSpecMeasureV2[];

  let tbody: HTMLTableSectionElement;
  let selected = new Set();
  let scroll = 0;
  let contentRect = new DOMRectReadOnly(0, 0, 0, 0);
  let wrapperRect = new DOMRectReadOnly(0, 0, 0, 0);

  $: tableWidth = contentRect.width;
  $: wrapperWidth = wrapperRect.width;
</script>

<div
  class="wrapper relative"
  on:scroll={(e) => {
    console.log(e);
    scroll = e.currentTarget.scrollLeft;
  }}
  bind:contentRect={wrapperRect}
>
  <table bind:contentRect>
    <colgroup>
      <col style:width="{gutterWidth}px" />
      <col style:max-width="120px" />
      <col style:max-width="120px" />
      <col style:max-width="120px" />
    </colgroup>
    <thead class="sticky top-0 z-10">
      <tr class:insert={$insertIndex === -1}>
        <th class="!pl-5">
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
            <th>{header}</th>
          {/if}
        {/each}
      </tr>
    </thead>
    <tbody bind:this={tbody} class="relative">
      {#each items as item, i (item.name)}
        <MetricsTableRow
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
          type={dimensions ? "dimensions" : "measures"}
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
</div>

<style lang="postcss">
  thead tr {
    height: 40px;
  }

  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-full;
    @apply font-normal cursor-pointer select-none relative;
  }

  .wrapper {
    @apply overflow-auto w-full max-w-full h-fit max-h-full relative bg-white;
    @apply border-[1px] rounded-[2px];
    @apply max-h-72;
  }

  th {
    @apply text-left;
    @apply pl-4 text-slate-500 bg-background;
    @apply border-b;
  }

  .insert th {
    @apply border-b border-primary-500;
  }
</style>
