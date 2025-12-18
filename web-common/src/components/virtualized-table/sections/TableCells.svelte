<script lang="ts">
  import type { VirtualItem } from "@tanstack/svelte-virtual";
  import type { VirtualizedTableColumns } from "../types";
  import Cell from "../core/Cell.svelte";
  import Row from "../core/Row.svelte";

  export let virtualColumnItems: VirtualItem[];
  export let virtualRowItems: VirtualItem[];
  export let rows;
  export let selectedIndex: number[] = [];
  export let columns: VirtualizedTableColumns[];
  export let scrolling = false;
  export let activeIndex: number;
  export let excludeMode = false;
  export let cellLabel: string | undefined = undefined;

  $: atLeastOneSelected = !!selectedIndex?.length;

  const getCellProps = (
    virtRow: VirtualItem,
    virtCol: VirtualItem,
    selectedIndex: number[],
  ) => {
    const column = columns[virtCol.index];
    const columnName = column.name;
    const value = rows[virtRow.index][columnName];
    return {
      value,
      formattedValue: rows[virtRow.index]["__formatted_" + columnName],
      type: column.type,
      barValue: column.max ? value / column.max : 0,
      rowSelected: selectedIndex.findIndex((tgt) => virtRow.index === tgt) >= 0,
      colSelected: column.highlight,
    };
  };
</script>

{#each virtualColumnItems as column (column.key)}
  <Row>
    {#each virtualRowItems as row (`${row.key}-${column.key}`)}
      {@const rowActive = activeIndex === row.index}
      <Cell
        {row}
        {column}
        {atLeastOneSelected}
        {excludeMode}
        {rowActive}
        suppressTooltip={scrolling}
        {...getCellProps(row, column, selectedIndex)}
        on:inspect
        on:select-item
        label={cellLabel}
      />
    {/each}
  </Row>
{/each}
