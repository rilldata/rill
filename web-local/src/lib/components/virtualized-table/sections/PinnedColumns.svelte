<script lang="ts">
  import type { VirtualizedTableColumns } from "../../../types";
  import { createEventDispatcher } from "svelte";
  import Cell from "../core/Cell.svelte";
  import ColumnHeader from "../core/ColumnHeader.svelte";
  import Row from "../core/Row.svelte";
  import type { PinnedColumnSide } from "../types";

  const dispatch = createEventDispatcher();

  export let side: PinnedColumnSide = "right";
  export let virtualRowItems;
  export let virtualColumnItems;
  export let pinnedColumns: VirtualizedTableColumns[];
  export let scrolling = false;
  export let activeIndex: number;
  export let rows;

  function reconcileVirtualColumns(
    virtualColumns,
    pinnedColumns: VirtualizedTableColumns[]
  ) {
    // for each pinned column, we need to add size + start (based on 0);
    let virtualColumnSet = pinnedColumns.map((columnProfile) => {
      let virtualColumn = virtualColumns.find(
        (column) => column.key === columnProfile.name
      );
      return {
        virtualColumn: { ...virtualColumn },
        columnProfile,
      };
    });
    let runningStart = 0;
    virtualColumnSet.forEach((columnInfo) => {
      // reset virtualColumn.start
      columnInfo.virtualColumn.start = runningStart;
      runningStart += columnInfo.virtualColumn.size;
    });
    return virtualColumnSet;
  }

  $: reconciledColumns = reconcileVirtualColumns(
    virtualColumnItems,
    pinnedColumns
  );

  $: totalWidth = reconciledColumns.reduce((total, column) => {
    total += column.virtualColumn.size;
    return total;
  }, 0);
</script>

<div
  style:right={side === "right" ? 0 : "auto"}
  style:left={side === "left" ? 0 : "auto"}
  class=" top-0 sticky z-40 border-l-2 border-gray-400"
  style:width="{totalWidth}px"
>
  <div class="w-full sticky relative top-0 z-10">
    {#each reconciledColumns as { columnProfile, virtualColumn }, i (columnProfile.name + "-pinned")}
      <ColumnHeader
        header={{
          start: virtualColumn.start,
          size: virtualColumn.size,
        }}
        enableResize={false}
        name={columnProfile.name}
        type={columnProfile.type}
        on:pin={() => dispatch("pin", columnProfile)}
        pinned={true}
      />
    {/each}
  </div>
  {#each reconciledColumns as { columnProfile, virtualColumn }, i (columnProfile.name + "-pinned")}
    <Row>
      {#each virtualRowItems as row (`${row.key}-${i}-pinned`)}
        {@const value = rows[row.index][columnProfile.name]}
        {@const type = columnProfile.type}
        {@const rowActive = activeIndex === row?.index}
        {@const suppressTooltip = scrolling}

        <Cell
          {suppressTooltip}
          {rowActive}
          {value}
          {row}
          column={{ start: virtualColumn.start, size: virtualColumn.size }}
          {type}
          on:inspect
        />
      {/each}
    </Row>
  {/each}
</div>
