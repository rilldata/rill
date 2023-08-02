<script lang="ts">
  import { createVirtualizer, VirtualItem } from "@tanstack/svelte-virtual";
  import PivotVirtualRow from "./PivotVirtualRow.svelte";
  import PivotCell from "./PivotCell.svelte";

  export let rowCt: number;
  export let colCt: number;
  export let fixedColCt: number;
  export let getColumnWidth: (idx: number) => number;
  export let getRowSize: (idx: number) => number;
  export let renderCell: (rowIdx: number, colIdx: number) => any;
  export let height: number;

  function range(n: number) {
    return new Array(n).fill(0).map((d, i) => i);
  }

  let container: HTMLDivElement;
  $: rowVirtualizer = createVirtualizer({
    count: rowCt,
    getScrollElement: () => container,
    estimateSize: getRowSize,
    overscan: 10,
  });

  let virtualRows: VirtualItem[] = [];
  let totalVerticalSize = 0;
  let paddingTop = 0;
  let paddingBottom = 0;
  $: {
    virtualRows = $rowVirtualizer?.getVirtualItems() ?? [];
    totalVerticalSize = $rowVirtualizer?.getTotalSize() ?? 0;
    paddingTop = virtualRows?.length > 0 ? virtualRows?.[0]?.start || 0 : 0;
    paddingBottom =
      virtualRows?.length > 0
        ? totalVerticalSize - (virtualRows?.at(-1)?.end || 0)
        : 0;
  }

  $: columnVirtualizer = createVirtualizer({
    horizontal: true,
    count: colCt,
    getScrollElement: () => container,
    estimateSize: getColumnWidth,
    overscan: 5,
  });

  let columnsToRender: VirtualItem[] = [];
  let totalHorizontalSize = 0;
  let paddingLeft = 0;
  let paddingRight = 0;
  $: {
    const virtualColumns = $columnVirtualizer?.getVirtualItems() ?? [];
    totalHorizontalSize = $columnVirtualizer?.getTotalSize() ?? 0;

    // Manually calculate fixed virtual column set as they may not be in current virtualized item set
    const fixedVirtualColumns: VirtualItem[] = range(fixedColCt).reduce(
      (arr, idx) => {
        const start = arr[idx - 1]?.end ?? 0;
        const size = getColumnWidth(idx);
        const end = start + getColumnWidth(idx);
        return [
          ...arr,
          {
            index: idx,
            start,
            end,
            key: idx,
            lane: 0,
            size,
          },
        ];
      },
      []
    );

    // If current virtual column set has fixed columns, remove them since we will use our measurements
    const virtualColumnsSansFixed = virtualColumns.filter(
      (c) => c.index >= fixedColCt
    );

    // Merge the fixed column set with the remaining virtual column set
    columnsToRender = [...fixedVirtualColumns, ...virtualColumnsSansFixed];

    paddingLeft =
      virtualColumns?.length > 0 ? virtualColumns?.[0]?.start || 0 : 0;
    // Adjust padding left for fixed virtual columns if needed
    paddingLeft = Math.max(0, paddingLeft - fixedVirtualColumns.at(-1).end);
    paddingRight =
      virtualColumns?.length > 0
        ? totalHorizontalSize - (virtualColumns?.at(-1)?.end || 0)
        : 0;
  }

  const isFixedColumn = (idx: number) => idx < fixedColCt;
</script>

<div
  bind:this={container}
  style={`height: ${height}px; overflow: auto;`}
  class="border"
>
  <table
    class="border-collapse"
    style={`height: ${totalVerticalSize}px; max-height: ${totalVerticalSize}px; width: ${totalHorizontalSize}px; max-width: ${totalHorizontalSize}px; overflow: none; table-layout: fixed`}
  >
    <thead class="sticky top-0 bg-gray-100 z-10">
      <PivotVirtualRow element="th" {paddingLeft} {paddingRight}>
        {#each columnsToRender as column (column.index)}
          <PivotCell
            rowIdx={-1}
            class={isFixedColumn(column.index)
              ? "bg-gray-200 text-left"
              : "text-left"}
            fixed={isFixedColumn(column.index)}
            element="th"
            {renderCell}
            item={column}
          />
        {/each}
      </PivotVirtualRow>
    </thead>
    <tbody>
      <!-- Virtual top padding row -->
      {#if paddingTop > 0}
        <tr>
          <td style={`height: ${paddingTop}px;`} />
        </tr>
      {/if}
      {#each virtualRows as row (row.index)}
        <PivotVirtualRow {paddingLeft} {paddingRight}>
          {#each columnsToRender as column (column.index)}
            <PivotCell
              rowIdx={row.index}
              rowHeight={getRowSize(row.index)}
              class={isFixedColumn(column.index) ? "bg-gray-100" : ""}
              fixed={isFixedColumn(column.index)}
              {renderCell}
              item={column}
            />
          {/each}
        </PivotVirtualRow>
      {/each}
      <!-- Virtual bottom padding row -->
      {#if paddingBottom > 0}
        <tr>
          <td style={`height: ${paddingBottom}px`} />
        </tr>
      {/if}
    </tbody>
  </table>
</div>
