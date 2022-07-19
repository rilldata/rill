<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

  import { dataModelerService } from "$lib/application-state-stores/application-store";

  import Button from "$lib/components/Button.svelte";

  import Close from "$lib/components/icons/Close.svelte";
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
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

  let metricsExplore: Readable<MetricsExploreEntity>;
  $: metricsExplore = getMetricsExploreById(metricsDefId);

  $: anythingSelected = isAnythingSelected($metricsExplore?.activeValues);
  function clearAllFilters() {
    clearSelectedLeaderboardValuesAndUpdate(store.dispatch, metricsDefId);
  }
  $: metricsDefinition = getMetricsDefReadableById(metricsDefId);
</script>

<header
  class="grid w-full bg-white self-stretch justify-between"
  style:grid-template-columns="auto auto"
  style:grid-template-rows="auto max-content"
>
  <div class="grid gap-y-2 grid-flow-row">
    <h1 style:line-height="1.1">
      <div class="pl-4 text-gray-600" style:font-size="24px">
        {#if $metricsDefinition}
          {$metricsDefinition?.metricDefLabel}
        {/if}
      </div>
    </h1>

    <div class="w-max self-start">
      <TimeRangeSelector {metricsDefId} />
    </div>
  </div>
  <div
    class="
    justify-items-end
    grid
    grid-flow-row
    h-max
  "
  >
    <Button
      type="secondary"
      on:click={() => {
        dataModelerService.dispatch("setActiveAsset", [
          EntityType.MetricsDefinition,
          metricsDefId,
        ]);
      }}
    >
      <div class="flex items-center gap-x-2">
        Edit Metrics <MetricsIcon />
      </div>
    </Button>

    <div class="justify-self-end self-start h-max">
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
  </div>
</header>
