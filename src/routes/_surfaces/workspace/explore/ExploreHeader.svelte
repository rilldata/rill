<script lang="ts">
  import { fly } from "svelte/transition";
  import Close from "$lib/components/icons/Close.svelte";
  import { store } from "$lib/redux-store/store-root";
  import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import { clearSelectedLeaderboardValuesAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import type { Readable } from "svelte/store";
  import { getMetricsExploreById } from "$lib/redux-store/explore/explore-readables";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";

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
  style:grid-template-columns="auto max-content"
  class="grid w-full bg-white self-stretch"
>
  <div>
    <h1 style:line-height="1.1">
      <div class="pl-2 text-gray-600" style:font-size="24px">
        {#if $metricsDefinition}
          {$metricsDefinition?.metricDefLabel}
        {/if}
      </div>
    </h1>
  </div>
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
</header>
