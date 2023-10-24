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
  export let headerStyle = "";
  export let bodyStyle = "";

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

  let fixedColumnsToRender: VirtualItem[] = [];
  let nonFixedColumnsToRender: VirtualItem[] = [];
  let columnsToRender: VirtualItem[] = [];
  $: {
    const virtualColumns = $columnVirtualizer?.getVirtualItems() ?? [];

    // Manually calculate fixed virtual column set as they may not be in current virtualized item set

    fixedColumnsToRender = range(fixedColCt).reduce(
      (arr: VirtualItem[], idx) => {
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
    nonFixedColumnsToRender = virtualColumns.filter(
      (c) => c.index >= fixedColCt
    );

    columnsToRender = fixedColumnsToRender.concat(nonFixedColumnsToRender);
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

  // Sync header scroll with body virtual scroll
  let scrollLeft = 0;
  const handleScroll = (evt) => {
    scrollLeft = evt.target.scrollLeft;
  };
</script>

<div role="grid" class="overflow-hidden" style={headerStyle}>
  <div
    role="row"
    class="sticky z-10 top-0 left-0 flex"
    style={`height: ${headerHeight}px;`}
  >
    <div role="presentation" class="flex">
      {#each fixedColumnsToRender as col (col.index)}
        <div
          role="cell"
          style={getCellWrapperStyle({ size: headerHeight }, col)}
        >
          <svelte:component
            this={headerComponent}
            colIdx={col.index}
            rowIdx={-1}
            fixed={true}
            lastFixed={col.index === fixedColCt - 1}
          />
        </div>
      {/each}
    </div>
    <div
      role="presentation"
      class="absolute left-0"
      style={`transform: translate3d(${-scrollLeft}px, 0px, 0px)`}
    >
      {#each nonFixedColumnsToRender as col (col.index)}
        <div
          role="cell"
          style={getCellWrapperStyle({ size: headerHeight }, col)}
        >
          <svelte:component
            this={headerComponent}
            colIdx={col.index}
            rowIdx={-1}
            fixed={false}
            lastFixed={false}
          />
        </div>
      {/each}
    </div>
  </div>
</div>
<div
  bind:this={container}
  role="grid"
  class="overflow-auto"
  style={`height: ${height}px; ${bodyStyle}`}
  on:scroll={handleScroll}
>
  <div
    class="relative"
    style={`height: ${$rowVirtualizer?.getTotalSize()}px; width: ${$columnVirtualizer?.getTotalSize()}px;`}
  >
    {#each $rowVirtualizer.getVirtualItems() as row (row.index)}
      <div
        role="row"
        class="absolute left-0 w-full flex"
        style={`height: ${row.size}px; transform: translateY(${row.start}px);`}
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
