<script lang="ts">
  import {
    getVisibleDimensionColor,
    toggleVisibleDimensions,
  } from "./time-dimension-details-store";
  import { useTDDContext } from "./context";
  import type { TCellData } from "./mock-data";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import { FILTER_OVERFLOW_WIDTH } from "./constants";

  export let cell: TCellData;
  export let isInFilter: boolean;

  let _class = "w-full h-full flex items-center justify-between";
  const { store } = useTDDContext();

  let dotClass = "";
  $: {
    const isVisibleOnChart =
      cell.text && $store.visibleDimensions.includes(cell.text);
    dotClass = `rounded-lg h-[7px] w-[7px] outline-offset-1 outline-1 outline`;
    const activeColor =
      cell.text && getVisibleDimensionColor($store, cell.text);
    if (isVisibleOnChart)
      dotClass += ` bg-${activeColor} outline-${activeColor}`;
    if (!isVisibleOnChart)
      dotClass += ` outline-gray-300 group-hover:outline-gray-500`;

    // Colors for Tailwind
    // bg-pink-600 bg-cyan-600 bg-green-600 bg-orange-600 bg-purple-600 bg-red-600 bg-blue-600
    // outline-pink-600 outline-cyan-600 outline-green-600 outline-orange-600 outline-purple-600 outline-red-600 outline-blue-600
  }

  const handleDotClick = () => {
    store.update((state) => {
      if (cell.text) toggleVisibleDimensions(state, cell.text);
      return state;
    });
  };
  // Position the filter check/X outside of the dimension value cell
  const filterOverflowStyle = `left: -${FILTER_OVERFLOW_WIDTH}px; width: ${FILTER_OVERFLOW_WIDTH}px;`;
</script>

<div class={_class}>
  <div
    style={filterOverflowStyle}
    class="absolute top-0 bg-white h-full flex items-center justify-center z-10"
  >
    {#if isInFilter}
      <Check size="20px" />
    {/if}
  </div>
  {cell.text}
  <button
    class="h-full w-4 flex items-center justify-center cursor-pointer group"
    on:click|preventDefault|stopPropagation={handleDotClick}
  >
    <div class={dotClass} />
  </button>
</div>
