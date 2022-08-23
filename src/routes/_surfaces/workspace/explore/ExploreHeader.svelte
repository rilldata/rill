<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import Button from "$lib/components/Button.svelte";
  import Close from "$lib/components/icons/Close.svelte";
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import { isFiltered } from "$lib/util/isFiltered";
  import { fly } from "svelte/transition";
  import TimeControls from "./time-controls/TimeControls.svelte";
  import {
    getMetricViewMetadata,
    getMetricViewMetaQueryKey,
  } from "$lib/svelte-query/queries/metric-view";
  import { useQuery, useQueryClient } from "@sveltestack/svelte-query";
  import { metricsExplorerStore } from "$lib/application-state-stores/explorer-stores";
  import { MetricViewMetaResponse } from "$common/rill-developer-service/MetricViewActions";
  import { invalidateMetricViewData } from "$lib/svelte-query/queries/metric-view";

  export let metricsDefId: string;

  const queryClient = useQueryClient();

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];

  let queryKey = getMetricViewMetaQueryKey(metricsDefId);
  const queryResult = useQuery<MetricViewMetaResponse, Error>(queryKey, () =>
    getMetricViewMetadata(metricsDefId)
  );
  $: {
    queryKey = getMetricViewMetaQueryKey(metricsDefId);
    queryResult.setOptions(queryKey, () => getMetricViewMetadata(metricsDefId));
  }
  $: metricsExplorerStore.sync(metricsDefId, $queryResult.data);

  $: filtered = isFiltered(metricsExplorer?.filters);
  function clearAllFilters() {
    metricsExplorerStore.clearFilters(metricsDefId);
    invalidateMetricViewData(queryClient, metricsDefId);
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
        {#if filtered}
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
