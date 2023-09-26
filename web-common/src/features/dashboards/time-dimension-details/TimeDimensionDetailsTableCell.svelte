<script lang="ts">
  import { createQuery, CreateQueryResult } from "@tanstack/svelte-query";
  import { getBlock } from "./util";
  import { fetchData, TCellData } from "./mock-data";
  import type { SvelteComponent } from "svelte";
  import { useTDDContext } from "./context";
  import { toggleFilter } from "./time-dimension-details-store";
  import { getCellComponent } from "./cell-renderings";

  export let rowIdx: number;
  export let colIdx: number;
  export let fixed = false;
  export let lastFixed = false;

  const { store } = useTDDContext();
  // If the current data block has this cell, get the data. Otherwise for now assume "loading" state (can handle errors later)
  let cellData: TCellData & { isLoading?: boolean } = {
    isLoading: true,
    text: "...",
  };
  let block = getBlock(100, rowIdx, rowIdx);
  $: block = getBlock(100, rowIdx, rowIdx);

  let rowDimension = "";
  $: isTableFiltered = $store.filteredValues.length > 0;
  $: isCellInFilter = $store.filteredValues.includes(rowDimension);

  const cellQuery = createQuery({
    queryKey: ["time-dimension-details", block[0], block[1]],
    queryFn: fetchData(block, 1000),
  }) as CreateQueryResult<{
    block: number[];
    data: { text?: string; value?: number; sparkline?: number[] }[][];
  }>;

  $: {
    if (
      $cellQuery.data &&
      rowIdx >= $cellQuery.data.block[0] &&
      rowIdx < $cellQuery.data.block[1]
    ) {
      cellData =
        $cellQuery.data.data[rowIdx - $cellQuery.data.block[0]][colIdx];
      rowDimension =
        $cellQuery.data.data[rowIdx - $cellQuery.data.block[0]][0].text;
    } else cellData = { text: "...", isLoading: true };
  }

  let _class = "";
  $: {
    _class = "h-full w-full flex items-center px-2";
    if (fixed) _class += ` z-10`;
    if (lastFixed) _class += ` right-shadow`;

    // Determine background color based on store
    const isRowHighlighted = $store.highlightedRow === rowIdx;
    const isColHighlighted = $store.highlightedCol === colIdx;
    const isHighlighted = isRowHighlighted || isColHighlighted;
    const isDoubleHighlighted = isRowHighlighted && isColHighlighted;
    const isScrubbed =
      $store.scrubbedCols &&
      colIdx >= $store.scrubbedCols.at(0) &&
      colIdx <= $store.scrubbedCols.at(1);

    let bgColors = {
      fixed: {
        base: "bg-slate-50",
        highlighted: "bg-slate-100",
        doubleHighlighted: "bg-slate-200",
      },
      scrubbed: {
        base: "bg-blue-50",
        highlighted: "bg-blue-100",
        doubleHighlighted: "bg-blue-200",
      },
      default: {
        base: "bg-white",
        highlighted: "bg-gray-100",
        doubleHighlighted: "bg-gray-200",
      },
    };

    // Choose palette based on type of cell state
    let palette = bgColors.default;
    if (fixed && colIdx !== 0) palette = bgColors.fixed;
    else if (isScrubbed) palette = bgColors.scrubbed;

    // Choose color within palette based on highlighted state
    let colorName = palette.base;
    if (isDoubleHighlighted) colorName = palette.doubleHighlighted;
    else if (isHighlighted) colorName = palette.highlighted;
    _class += ` ${colorName}`;

    // Filter states
    if (isTableFiltered && !isCellInFilter) _class += ` ui-copy-disabled-faint`;
  }

  const handleMouseEnter = () => {
    $store.highlightedRow = rowIdx;
    $store.highlightedCol = colIdx;
  };
  const handleMouseLeave = () => {
    $store.highlightedCol = null;
    $store.highlightedRow = null;
  };

  // TODO: with real data, this should be dependent on selected metric
  const format = "0.2f";
  let cellComponent: typeof SvelteComponent<any>;
  let cellComponentDefaultProps = {};
  $: {
    ({ cellComponent, cellComponentDefaultProps } = getCellComponent(
      colIdx,
      format
    ));
  }
  const handleClick = () => {
    store.update((state) => {
      toggleFilter(state, rowDimension);
      return state;
    });
  };
</script>

<button
  class={_class}
  on:mouseenter={handleMouseEnter}
  on:mouseleave={handleMouseLeave}
  on:click={handleClick}
>
  {#if cellComponent && !cellData.isLoading}
    <svelte:component
      this={cellComponent}
      {...cellComponentDefaultProps}
      cell={cellData}
      isInFilter={isTableFiltered && isCellInFilter}
      isOutOfFilter={isTableFiltered && !isCellInFilter}
    />
  {:else}
    {cellData?.text ?? cellData?.value}
  {/if}
</button>

<style>
  .right-shadow::after {
    content: "";
    width: 1px;
    height: 100%;
    position: absolute;
    top: 0px;
    right: 0px;
    background: #e5e7eb;
    filter: drop-shadow(3px 0px 3px rgb(0 0 0 / 0.27));
  }
</style>
