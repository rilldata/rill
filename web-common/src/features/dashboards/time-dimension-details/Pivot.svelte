<script lang="ts">
  import { createVirtualizer, VirtualItem } from "@tanstack/svelte-virtual";
  import type { SvelteComponent } from "svelte";

  export let rowCt: number;
  export let colCt: number;
  export let fixedColCt: number;
  export let getColumnWidth: (idx: number) => number;
  export let getRowSize: (idx: number) => number;
  export let cellComponent: typeof SvelteComponent;
  export let headerComponent: typeof SvelteComponent;
  export let height: number;
  export let headerHeight: number;

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

  $: columnVirtualizer = createVirtualizer({
    horizontal: true,
    count: colCt,
    getScrollElement: () => container,
    estimateSize: getColumnWidth,
    overscan: 10,
  });

  let columnsToRender: VirtualItem[] = [];
  $: {
    const virtualColumns = $columnVirtualizer?.getVirtualItems() ?? [];

    // Manually calculate fixed virtual column set as they may not be in current virtualized item set
    const fixedVirtualColumns = range(fixedColCt).reduce((arr, idx) => {
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
    const nonfixedVirtualColumns = virtualColumns.filter(
      (c) => c.index >= fixedColCt
    );

    columnsToRender = fixedVirtualColumns.concat(nonfixedVirtualColumns);
  }

  const isFixedColumn = (idx: number) => idx < fixedColCt;
  const getCellWrapperStyle = (row: Partial<VirtualItem>, col: VirtualItem) => {
    let style = `display: inline-block; width: ${col.size}px; height: ${row.size}px;`;
    if (isFixedColumn(col.index))
      style += `position: sticky; left: ${col.start}px; top: 0px; z-index: 2; transform: translateX(0px);`;
    else
      style += `position: absolute; top: 0; left: 0;  transform: translateX(${col.start}px);`;
    return style;
  };
</script>

<div
  bind:this={container}
  role="grid"
  style={`height: ${height}px; overflow: auto;`}
  class="border"
>
  <div
    style={`height: ${$rowVirtualizer?.getTotalSize()}px; width: ${$columnVirtualizer?.getTotalSize()}px; position: relative;`}
  >
    <div
      role="row"
      style={`position: sticky; z-index: 2; top: 0; left: 0; height: ${headerHeight}px; display: flex;`}
    >
      {#each columnsToRender as col (col.index)}
        <div
          role="cell"
          style={getCellWrapperStyle({ size: headerHeight }, col)}
        >
          <svelte:component
            this={headerComponent}
            colIdx={col.index}
            rowIdx={-1}
            fixed={isFixedColumn(col.index)}
            lastFixed={col.index === fixedColCt - 1}
          />
        </div>
      {/each}
    </div>
    {#each $rowVirtualizer.getVirtualItems() as row (row.index)}
      <div
        role="row"
        style={`position: absolute; top: ${headerHeight}; left: 0; width: 100%; height: ${row.size}px; transform: translateY(${row.start}px);`}
      >
        {#each columnsToRender as col (col.index)}
          <div role="cell" style={getCellWrapperStyle(row, col)}>
            <svelte:component
              this={cellComponent}
              colIdx={col.index}
              rowIdx={row.index}
              fixed={isFixedColumn(col.index)}
              lastFixed={col.index === fixedColCt - 1}
            />
          </div>
        {/each}
      </div>
    {/each}
  </div>
</div>
