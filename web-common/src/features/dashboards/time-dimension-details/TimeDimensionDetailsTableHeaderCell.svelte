<script lang="ts">
  import type { Writable } from "svelte/store";

  export let rowIdx: number;
  export let colIdx: number;
  export let fixed: boolean;
  export let store: Writable<{
    highlightedCol: number | null;
    highlightedRow: number | null;
    scrubbedCols: [number, number] | null;
  }>;

  let _class = "";
  $: {
    _class = "h-full bg-white border-b text-left";
    const isRowHighlighted = $store.highlightedRow === rowIdx;
    const isColHighlighted = $store.highlightedCol === colIdx;
    const isHighlighted = isRowHighlighted || isColHighlighted;
    const isDoubleHighlighted = isRowHighlighted && isColHighlighted;
    const isScrubbed =
      $store.scrubbedCols &&
      colIdx >= $store.scrubbedCols.at(0) &&
      colIdx <= $store.scrubbedCols.at(1);

   _class += ` `;

    if (fixed) _class += ` z-2`;
    if(colIdx === 5) _class += ` right-shadow`
  }

  const handleMouseEnter = () => {
    $store.highlightedCol = colIdx;
  };
  const handleMouseLeave = () => {
    $store.highlightedCol = null;
  };
</script>


  <div
    class={_class}
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
  >
    Col {colIdx}
  </div>

  <style>

    .right-shadow:after {
        content: "";
        width: 1px;
        height: 100%;
        position: absolute;
        top: 0px;
        right: 0px;
        background: #e5e7eb;
        filter: drop-shadow(3px 0px 3px rgb(0 0 0 / 0.27))
    }
</style>