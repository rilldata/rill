<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import Button from "$lib/components/Button.svelte";
  import Close from "$lib/components/icons/Close.svelte";
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import { clearSelectedLeaderboardValuesAndUpdate } from "$lib/redux-store/explore/explore-apis";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { store } from "$lib/redux-store/store-root";
  import { isAnythingSelected } from "$lib/util/isAnythingSelected";
  import type { Readable } from "svelte/store";
  import { fly } from "svelte/transition";
  import TimeControls from "./time-controls/TimeControls.svelte";
  import { invalidateMetricViewTopList } from "$lib/svelte-query/queries/metric-view";
  import { useQueryClient } from "@sveltestack/svelte-query";

  export let metricsDefId: string;

  const queryClient = useQueryClient();

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  $: anythingSelected = isAnythingSelected($metricsExplorer?.filters);
  function clearAllFilters() {
    clearSelectedLeaderboardValuesAndUpdate(store.dispatch, metricsDefId);
    invalidateMetricViewTopList(queryClient, metricsDefId);
  }
  $: metricsDefinition = getMetricsDefReadableById(metricsDefId);
</script>

<header
  class="grid w-full bg-white self-stretch justify-between px-4 pt-3"
  style:grid-template-columns="auto auto"
  style:grid-template-rows="auto max-content"
>
  <div class="grid gap-y-2 grid-flow-row">
    <h1 style:line-height="1.1" class="pt-3">
      <div class="pl-4 text-gray-600" style:font-size="24px">
        {#if $metricsDefinition}
          {$metricsDefinition?.metricDefLabel}
        {/if}
      </div>
    </h1>

    <TimeControls {metricsDefId} />
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
