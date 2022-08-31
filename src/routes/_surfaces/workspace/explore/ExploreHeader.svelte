<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import { Button } from "$lib/components/button";
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import Filters from "./filters/Filters.svelte";
  import TimeControls from "./time-controls/TimeControls.svelte";

  export let metricsDefId: string;

  $: metricsDefinition = getMetricsDefReadableById(metricsDefId);
</script>

<section id="header" class="w-full flex flex-col">
  <!-- top row
    title and call to action
  -->
  <div class="flex justify-between w-full pt-4 pl-1 pr-4">
    <!-- title element -->
    <h1 style:line-height="1.1">
      <div class="pl-4 text-gray-700" style:font-size="24px">
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
  <div class="px-2">
    <TimeControls {metricsDefId} />
    {#key metricsDefId}
      <Filters {metricsDefId} />
    {/key}
  </div>
</section>
