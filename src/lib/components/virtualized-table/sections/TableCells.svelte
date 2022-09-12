<script lang="ts">
  import type { VirtualizedTableColumns } from "$lib/types";
  import Cell from "../core/Cell.svelte";
  import Row from "../core/Row.svelte";

  export let virtualColumnItems;
  export let virtualRowItems;
  export let rows;
  export let selectedIndex = [];
  export let columns: VirtualizedTableColumns[];
  export let scrolling = false;
  export let activeIndex: number;

  $: atLeastOneSelected = !!selectedIndex?.length;
</script>

{#each virtualColumnItems as column (column.key)}
  <Row>
    {#each virtualRowItems as row (`${row.key}-${column.key}`)}
      {@const formattedValue =
        rows[row.index]["__formatted_" + columns[column.index]?.name]}
      {@const value = rows[row.index][columns[column.index]?.name]}
      {@const type = columns[column.index]?.type}
      {@const rowActive = activeIndex === row?.index}
      {@const suppressTooltip = scrolling}
      {@const rowSelected =
        selectedIndex.findIndex((value) => row?.index === value) >= 0}
      <Cell
        {suppressTooltip}
        {rowActive}
        {value}
        {formattedValue}
        {row}
        {column}
        {type}
        {rowSelected}
        {atLeastOneSelected}
        on:inspect
        on:select-item
      />
    {/each}
  </Row>
{/each}
