<script lang="ts">
  import type { Writable } from "svelte/store";
  import type { TimeDimensionDetailsStore } from "./time-dimension-details-store";

  export let rowIdx: number;
  export let colIdx: number;
  export let fixed: boolean;
  export let lastFixed: boolean;
  export let store: Writable<TimeDimensionDetailsStore>;

  let _class = "";
  $: {
    _class = "h-full ";
    if (fixed) _class += ` z-2`;
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
    if (fixed) palette = bgColors.fixed;
    else if (isScrubbed) palette = bgColors.scrubbed;

    // Choose color within palette based on highlighted state
    let colorName = palette.base;
    if (isDoubleHighlighted) colorName = palette.doubleHighlighted;
    else if (isHighlighted) colorName = palette.highlighted;
    _class += ` ${colorName}`;
  }

  const handleMouseEnter = () => {
    $store.highlightedRow = rowIdx;
    $store.highlightedCol = colIdx;
  };
  const handleMouseLeave = () => {
    $store.highlightedCol = null;
    $store.highlightedRow = null;
  };
</script>

{#if rowIdx === -1}
  <div
    class={_class}
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
  >
    Column {colIdx}
  </div>
{:else}
  <div
    class={_class}
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
  >
    cell {rowIdx},{colIdx}
  </div>
{/if}

<style>
  .right-shadow:after {
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
