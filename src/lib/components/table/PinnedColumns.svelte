<script lang="ts">
  import { config } from "./config";

  import type { ProfileColumn } from "$lib/types";
  import { createEventDispatcher } from "svelte";
  import Cell from "./Cell.svelte";
  import ColumnHeader from "./ColumnHeader.svelte";
  import Row from "./Row.svelte";

  const dispatch = createEventDispatcher();

  export let virtualRowItems;
  export let pinnedColumns: ProfileColumn[];
  export let scrolling = false;
  export let activeIndex: number;
  export let rows;
</script>

<div
  style:right={0}
  class=" top-0 sticky z-40 border-l-2 border-gray-400"
  style:width="{pinnedColumns.length * config.columnWidth}px"
>
  <div class="w-full sticky relative top-0 z-10">
    {#each pinnedColumns as column, i (column.name)}
      <ColumnHeader
        header={{ start: i * config.columnWidth, size: config.columnWidth }}
        name={column.name}
        type={column.type}
        on:pin={() => dispatch("pin", column)}
        pinned={true}
      />
    {/each}
  </div>
  {#each pinnedColumns as column, i (column.name)}
    <Row>
      {#each virtualRowItems as row (`${row.key}-${i}`)}
        {@const value = rows[row.index][column.name]}
        {@const type = column.type}
        {@const rowActive = activeIndex === row?.index}
        {@const suppressTooltip = scrolling}

        <Cell
          {suppressTooltip}
          {rowActive}
          {value}
          {row}
          column={{ start: i * config.columnWidth, size: config.columnWidth }}
          {type}
          on:inspect
        />
      {/each}
    </Row>
  {/each}
</div>
