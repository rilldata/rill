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
  export let renderHeaderCell: (rowIdx: number, colIdx: number) => any;
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

  let totalHorizontalSize = 0;
  let paddingLeft = 0;
  let paddingRight = 0;
  let fixedVirtualColumns: VirtualItem[] = [];
  let nonfixedVirtualColumns: VirtualItem[] = [];
  $: {
    const virtualColumns = $columnVirtualizer?.getVirtualItems() ?? [];
    totalHorizontalSize = $columnVirtualizer?.getTotalSize() ?? 0;

    // Manually calculate fixed virtual column set as they may not be in current virtualized item set
    fixedVirtualColumns = range(fixedColCt).reduce((arr, idx) => {
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
    }, []);

    // If current virtual column set has fixed columns, remove them since we will use our measurements
    nonfixedVirtualColumns = virtualColumns.filter(
      (c) => c.index >= fixedColCt
    );

    const fullPaddingLeft =
      virtualColumns?.length > 0 ? virtualColumns?.[0]?.start || 0 : 0;
    // Adjust padding left for fixed virtual columns if needed
    paddingLeft = Math.max(0, fullPaddingLeft - fixedVirtualColumns.at(-1).end);
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
    <thead class="sticky top-0 z-10">
      <PivotVirtualRow element="th" {paddingLeft} {paddingRight}>
        <svelte:fragment slot="pre">
          {#each fixedVirtualColumns as column (column.index)}
            <PivotCell
              rowIdx={-1}
              fixed={isFixedColumn(column.index)}
              element="th"
              renderCell={renderHeaderCell}
              item={column}
            />
          {/each}
        </svelte:fragment>
        <svelte:fragment slot="body">
          {#each nonfixedVirtualColumns as column (column.index)}
            <PivotCell
              rowIdx={-1}
              fixed={isFixedColumn(column.index)}
              element="th"
              renderCell={renderHeaderCell}
              item={column}
            />
          {/each}
        </svelte:fragment>
      </PivotVirtualRow>
    </thead>
    <tbody>
      <!-- Virtual top padding row -->
      {#if paddingTop > 0}
        <PivotVirtualRow {paddingLeft} {paddingRight}>
          <svelte:fragment slot="pre">
            {#each fixedVirtualColumns as column (column.index)}
              <PivotCell
                rowIdx={-1}
                rowHeight={paddingTop}
                fixed
                renderCell={() => ""}
                item={column}
              />
            {/each}
          </svelte:fragment>
          <svelte:fragment slot="body">
            {#each nonfixedVirtualColumns as column (column.index)}
              <PivotCell
                rowIdx={-1}
                rowHeight={paddingTop}
                renderCell={() => ""}
                item={column}
              />
            {/each}
          </svelte:fragment>
        </PivotVirtualRow>
      {/if}
      {#each virtualRows as row (row.index)}
        <PivotVirtualRow {paddingLeft} {paddingRight}>
          <svelte:fragment slot="pre">
            {#each fixedVirtualColumns as column (column.index)}
              <PivotCell
                rowIdx={row.index}
                rowHeight={getRowSize(row.index)}
                fixed
                {renderCell}
                item={column}
              />
            {/each}
          </svelte:fragment>
          <svelte:fragment slot="body">
            {#each nonfixedVirtualColumns as column (column.index)}
              <PivotCell
                rowIdx={row.index}
                rowHeight={getRowSize(row.index)}
                {renderCell}
                item={column}
              />
            {/each}
          </svelte:fragment>
        </PivotVirtualRow>
      {/each}
      <!-- Virtual bottom padding row -->
      {#if paddingBottom > 0}
        <PivotVirtualRow {paddingLeft} {paddingRight}>
          <svelte:fragment slot="pre">
            {#each fixedVirtualColumns as column (column.index)}
              <PivotCell
                rowIdx={-1}
                rowHeight={paddingBottom}
                fixed
                renderCell={() => ""}
                item={column}
              />
            {/each}
          </svelte:fragment>
          <svelte:fragment slot="body">
            {#each nonfixedVirtualColumns as column (column.index)}
              <PivotCell
                rowIdx={-1}
                rowHeight={paddingBottom}
                renderCell={() => ""}
                item={column}
              />
            {/each}
          </svelte:fragment>
        </PivotVirtualRow>
      {/if}
    </tbody>
  </table>
</div>
