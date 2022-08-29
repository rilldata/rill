<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import { metricsExplorerStore } from "$lib/application-state-stores/explorer-stores";
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import {
    invalidateMetricViewData,
    useGetMetricViewMeta,
  } from "$lib/svelte-query/queries/metric-view";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import Filters from "./filters/Filters.svelte";
  import TimeControls from "./time-controls/TimeControls.svelte";
  import { Button } from "$lib/components/button";

  export let metricsDefId: string;

  const queryClient = useQueryClient();

  $: metaQuery = useGetMetricViewMeta(metricsDefId);
  // TODO: move this "sync" to a more relevant component
  $: if (metricsDefId && $metaQuery && metricsDefId === $metaQuery.data.id) {
    if (
      !$metaQuery.data.measures?.length ||
      !$metaQuery.data.dimensions?.length
    ) {
      dataModelerService.dispatch("setActiveAsset", [
        EntityType.MetricsDefinition,
        metricsDefId,
      ]);
    } else {
      invalidateMetricViewData(queryClient, metricsDefId);
    }
    metricsExplorerStore.sync(metricsDefId, $metaQuery.data);
  }

  $: metricsDefinition = getMetricsDefReadableById(metricsDefId);
</script>

<section id="header" class="w-full flex flex-col gap-y-3">
  <!-- top row -->
  <div class="flex justify-between w-full pt-3 pl-4 pr-4">
    <!-- title element -->
    <h1 style:line-height="1.1" class="pt-3">
      <div class="pl-4 text-gray-600" style:font-size="24px">
        {#if $metricsDefinition}
          {$metricsDefinition?.metricDefLabel}
        {/if}
      </div>
    </h1>
    <!-- top right CTAs -->
    <div>
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
    </div>
  </div>
  <!-- bottom row -->
  <div class="px-4">
    <TimeControls {metricsDefId} />
    {#key metricsDefId}
      <Filters {metricsDefId} />
    {/key}
  </div>
</section>
