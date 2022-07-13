<script lang="ts">
  import Close from "$lib/components/icons/Close.svelte";
  import { clearSelectedLeaderboardValuesAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExploreById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import type { Readable } from "svelte/store";
  import { fly } from "svelte/transition";
  import TimeRangeSelector from "./TimeRangeSelector.svelte";

  export let metricsDefId: string;

  let metricsLeaderboard: Readable<MetricsExploreEntity>;
  $: metricsLeaderboard = getMetricsExploreById(metricsDefId);

  $: anythingSelected = isAnythingSelected($metricsLeaderboard?.activeValues);
  function clearAllFilters() {
    clearSelectedLeaderboardValuesAndUpdate(store.dispatch, metricsDefId);
  }
  $: metricsDefinition = getMetricsDefReadableById(metricsDefId);
</script>

<header
  class="grid gap-y-2 w-full bg-white self-stretch"
  style:grid-template-columns="auto max-content"
  style:grid-template-rows="auto max-content"
>
  <h1 style:line-height="1.1">
    <div class="pl-4 text-gray-600" style:font-size="24px">
      {#if $metricsDefinition}
        {$metricsDefinition?.metricDefLabel}
      {/if}
    </div>
  </h1>
  <div class="justify-self-end">
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
    <!-- NOTE: place share buttons here -->
  </div>
  <div>
    <TimeRangeSelector {metricsDefId} />
  </div>
</header>
