<script lang="ts">
  import Checkbox from "./Checkbox.svelte";
  import MetricsTableRow from "./MetricsTableRow.svelte";
  import { insertIndex } from "./MetricsTableRow.svelte";
  import { YAMLMap } from "yaml";

  const headers = ["Name", "Definition", "Label", "Format", "Description"];
  const gutterWidth = 56;

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
  style:max-height="{items?.length * 40}px"
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
    <tbody bind:this={tbody} class="relative overflow-hidden">
      {#each items as item, i (i)}
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
</div>

<style lang="postcss">
  .wrapper {
    @apply overflow-x-auto overflow-y-hidden w-full max-w-full relative;
    @apply border rounded-[2px] min-h-fit h-fit;
    /* @apply max-h-72; */
  }

  table {
    @apply p-0 m-0 border-spacing-0 border-separate w-full;
    @apply font-normal cursor-pointer select-none relative h-fit;
  }

  thead tr {
    height: 40px;
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
