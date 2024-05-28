<script lang="ts">
  import ColumnHeaders from "@rilldata/web-common/components/virtualized-table/sections/ColumnHeaders.svelte";
  import TableCells from "@rilldata/web-common/components/virtualized-table/sections/TableCells.svelte";
  import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
  import type { DimensionTableRow } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-types";
  import {
    estimateColumnCharacterWidths,
    estimateColumnSizes,
  } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-utils";
  import { DIMENSION_TABLE_CONFIG as config } from "@rilldata/web-common/features/dashboards/dimension-table/DimensionTableConfig";
  import { createVirtualizer } from "@tanstack/svelte-virtual";
  import { setContext } from "svelte";

  export let rows: DimensionTableRow[];
  export let columns: VirtualizedTableColumns[];

  /** the overscan values tell us how much to render off-screen. These may be set by the consumer
   * in certain circumstances. The tradeoff: the higher the overscan amount, the more DOM elements we have
   * to render on initial load.
   */
  export let rowOverscanAmount = 40;
  export let columnOverscanAmount = 5;

  let container: HTMLDivElement;
  let containerWidth: number;
  let estimateColumnSize: number[] = [];

  let rowScrollOffset = 0;
  $: rowScrollOffset = $rowVirtualizer?.scrollOffset || 0;
  let colScrollOffset = 0;
  $: colScrollOffset = $columnVirtualizer?.scrollOffset || 0;

  const { columnWidths } = estimateColumnCharacterWidths(columns, rows);

  /* set context for child components */
  setContext("config", config);

  $: rowVirtualizer = createVirtualizer({
    getScrollElement: () => container,
    count: rows.length,
    estimateSize: () => config.rowHeight,
    overscan: rowOverscanAmount,
    paddingStart: config.columnHeaderHeight,
    initialOffset: rowScrollOffset,
  });

  $: if (rows && columns) {
    estimateColumnSize = estimateColumnSizes(
      columns,
      columnWidths,
      containerWidth,
      config,
    );
  }

  $: columnVirtualizer = createVirtualizer({
    getScrollElement: () => container,
    horizontal: true,
    count: columns.length,
    getItemKey: (index) => columns[index].name,
    estimateSize: (index) => {
      return estimateColumnSize[index];
    },
    overscan: columnOverscanAmount,
    initialOffset: colScrollOffset,
  });

  $: virtualRows = $rowVirtualizer?.getVirtualItems() ?? [];
  $: virtualHeight = $rowVirtualizer?.getTotalSize() ?? 0;

  $: virtualColumns = $columnVirtualizer?.getVirtualItems() ?? [];
  $: virtualWidth = $columnVirtualizer?.getTotalSize() ?? 0;
</script>

<div
  bind:clientWidth={containerWidth}
  style="height: 100%;"
  role="table"
  aria-label="alert preview table"
>
  <div
    bind:this={container}
    style:width="100%"
    style:height="100%"
    class="overflow-auto grid max-w-fit"
    style:grid-template-columns="max-content auto"
  >
    <div
      role="grid"
      tabindex="0"
      class="relative surface"
      style:will-change="transform, contents"
      style:width="{virtualWidth}px"
      style:height="{virtualHeight}px"
    >
      <ColumnHeaders
        {columns}
        virtualColumnItems={virtualColumns}
        noPin
        showDataIcon={false}
      />
      {#if rows.length}
        <TableCells
          virtualColumnItems={virtualColumns}
          virtualRowItems={virtualRows}
          {columns}
          {rows}
          activeIndex={-1}
          cellLabel="Filter dimension value"
        />
      {/if}
    </div>
  </div>
</div>
