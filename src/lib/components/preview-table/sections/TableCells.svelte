<script lang="ts">
  import type { ProfileColumn } from "$lib/types";
  import Cell from "../core/Cell.svelte";
  import Row from "../core/Row.svelte";
  export let virtualColumnItems;
  export let virtualRowItems;
  export let rows;
  export let columns: ProfileColumn[];
  export let scrolling = false;
  export let activeIndex: number;
</script>

{#each virtualColumnItems as column (column.key)}
  <Row>
    {#each virtualRowItems as row (`${row.key}-${column.key}`)}
      {@const value = rows[row.index][columns[column.index]?.name]}
      {@const type = columns[column.index]?.type}
      {@const rowActive = activeIndex === row?.index}
      {@const suppressTooltip = scrolling}

      <Cell
        {suppressTooltip}
        {rowActive}
        {value}
        {row}
        {column}
        {type}
        on:inspect
      />
    {/each}
  </Row>
{/each}
