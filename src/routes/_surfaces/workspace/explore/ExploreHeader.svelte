<script lang="ts">
  import { fly } from "svelte/transition";
  import CheckerFull from "$lib/components/icons/CheckerFull.svelte";
  import CheckerHalf from "$lib/components/icons/CheckerHalf.svelte";
  import Close from "$lib/components/icons/Close.svelte";
  import Tooltip from "$lib/components/tooltip/Tooltip.svelte";
  import TooltipContent from "$lib/components/tooltip/TooltipContent.svelte";
  import { store } from "$lib/redux-store/store-root";
  import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import { clearSelectedLeaderboardValuesAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import type { Readable } from "svelte/store";
  import { getMetricsExploreById } from "$lib/redux-store/explore/explore-readables";

  export let metricsDefId: string;
  export let whichReferenceValue = "global";

  let metricsLeaderboard: Readable<MetricsExploreEntity>;
  $: metricsLeaderboard = getMetricsExploreById(metricsDefId);

  $: anythingSelected = isAnythingSelected($metricsLeaderboard?.activeValues);
  function clearAllFilters() {
    clearSelectedLeaderboardValuesAndUpdate(store.dispatch, metricsDefId);
  }
</script>

<header
  style:grid-template-columns="auto max-content"
  class="pb-6 pt-6 grid w-full bg-white"
>
  <div>
    <h1 style:line-height="1.1">
      <div class="pl-2 text-gray-600 font-normal" style:font-size="1.5rem">
        Total Records
      </div>
    </h1>
  </div>
  <div class="justify-self-end">
    <div
      style:font-size="24px"
      class="grid justify-items-end justify-end grid-flow-col items-center"
    >
      <Tooltip distance={16}>
        <button
          class="m-0 p-1 transition-color"
          class:bg-transparent={whichReferenceValue !== "filtered"}
          class:bg-gray-200={whichReferenceValue === "filtered"}
          class:font-bold={whichReferenceValue === "filtered"}
          class:text-gray-400={whichReferenceValue !== "filtered"}
          on:click={() => (whichReferenceValue = "filtered")}
          ><CheckerHalf /></button
        >
        <TooltipContent slot="tooltip-content">
          scale leaderboard bars by currently-filtered total
        </TooltipContent>
      </Tooltip>
      <Tooltip distance={16}>
        <button
          class="m-0 p-1 transition-color"
          class:bg-transparent={whichReferenceValue !== "global"}
          class:bg-gray-200={whichReferenceValue === "global"}
          class:font-bold={whichReferenceValue === "global"}
          class:text-gray-400={whichReferenceValue !== "global"}
          on:click={() => (whichReferenceValue = "global")}
          ><CheckerFull /></button
        >
        <TooltipContent slot="tooltip-content">
          scale leaderboard bars by total record count
        </TooltipContent>
      </Tooltip>
    </div>
    <div class="pt-3">
      {#if anythingSelected}
        <button
          transition:fly={{ duration: 200, y: 5 }}
          on:click={clearAllFilters}
          class="
                  grid gap-x-2 items-center font-bold
                  bg-red-100
                  text-red-900
                  p-1
                  pl-2 pr-2
                  rounded
              "
          style:grid-template-columns="auto max-content"
        >
          clear all filters <Close />
        </button>
      {/if}
    </div>
  </div>
</header>
