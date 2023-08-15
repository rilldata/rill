<script lang="ts">
  import {
    getVisibleDimensionColor,
    toggleVisibleDimensions,
  } from "./time-dimension-details-store";
  import { useTDTContext } from "./context";
  import type { TCellData } from "./mock-data";

  export let cell: TCellData;

  let _class = "w-full h-full flex items-center justify-between ";
  const { store } = useTDTContext();

  let dotClass = "";
  $: {
    const isVisibleOnChart = $store.visibleDimensions.includes(cell.text);
    dotClass = `rounded-lg h-[7px] w-[7px] outline-offset-1 outline-1 outline`;
    const activeColor = getVisibleDimensionColor($store, cell.text);
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
      toggleVisibleDimensions(state, cell.text);
      return state;
    });
  };
</script>

<div class={_class}>
  {cell.text}
  <button
    class="h-full w-4 flex items-center justify-center cursor-pointer group"
    on:click={handleDotClick}
  >
    <div class={dotClass} />
  </button>
</div>
