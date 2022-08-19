<script lang="ts">
  import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
  import { dataModelerService } from "$lib/application-state-stores/application-store";
  import Button from "$lib/components/Button.svelte";
  import MetricsIcon from "$lib/components/icons/Metrics.svelte";
  import { getMetricsExplorerById } from "$lib/redux-store/explore/explore-readables";
  import type { MetricsExplorerEntity } from "$lib/redux-store/explore/explore-slice";
  import { getMetricsDefReadableById } from "$lib/redux-store/metrics-definition/metrics-definition-readables";
  import type { Readable } from "svelte/store";
  import Filters from "./filters/Filters.svelte";
  import TimeControls from "./time-controls/TimeControls.svelte";

  export let metricsDefId: string;

  let metricsExplorer: Readable<MetricsExplorerEntity>;
  $: metricsExplorer = getMetricsExplorerById(metricsDefId);

  $: metricsDefinition = getMetricsDefReadableById(metricsDefId);
</script>

<section id="header" class="w-full flex flex-col gap-y-3">
  <!-- top row
    title and call to action
  -->
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
    <Filters {metricsDefId} values={$metricsExplorer?.activeValues} />
  </div>
</section>
