<script lang="ts">
  import type { VirtualizedTableColumns } from "@rilldata/web-local/lib/types";
  import Cell from "../core/Cell.svelte";
  import Row from "../core/Row.svelte";

  export let virtualColumnItems;
  export let virtualRowItems;
  export let rows;
  export let selectedIndex = [];
  export let columns: VirtualizedTableColumns[];
  export let scrolling = false;
  export let activeIndex: number;
  export let excludeMode = false;

  $: atLeastOneSelected = !!selectedIndex?.length;

  const getCellProps = (row, column) => {
    const value = rows[row.index][columns[column.index]?.name];
    return {
      value,
      formattedValue:
        rows[row.index]["__formatted_" + columns[column.index]?.name],
      type: columns[column.index]?.type,
      barValue: columns[column.index]?.total
        ? value / columns[column.index]?.total
        : 0,
      rowSelected: selectedIndex.findIndex((tgt) => row?.index === tgt) >= 0,
    };
  };
</script>

{#each virtualColumnItems as column (column.key)}
  <Row>
    {#each virtualRowItems as row (`${row.key}-${column.key}`)}
      {@const rowActive = activeIndex === row?.index}
      <Cell
        {row}
        {column}
        {atLeastOneSelected}
        {excludeMode}
        {rowActive}
        suppressTooltip={scrolling}
        {...getCellProps(row, column)}
        on:inspect
        on:select-item
      />
    {/each}
  </Row>
{/each}
