<script lang="ts">
  import type { MetricViewRequestFilter } from "$common/rill-developer-service/MetricViewActions";

  import {
    MetricsExplorerEntity,
    metricsExplorerStore,
  } from "$lib/application-state-stores/explorer-stores";
  import Close from "$lib/components/icons/Close.svelte";
  import { invalidateMetricViewData } from "$lib/svelte-query/queries/metric-view";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { fly } from "svelte/transition";
  export let metricsDefId;

  const queryClient = useQueryClient();

  let metricsExplorer: MetricsExplorerEntity;
  $: metricsExplorer = $metricsExplorerStore.entities[metricsDefId];

  function clearAllFilters() {
    metricsExplorerStore.clearFilters(metricsDefId);
    invalidateMetricViewData(queryClient, metricsDefId);
  }

  function isFiltered(filters: MetricViewRequestFilter): boolean {
    if (!filters) return false;
    return filters.include.length > 0 || filters.exclude.length > 0;
  }

  $: hasFilters = isFiltered(metricsExplorer?.filters);
</script>

<div class="pt-3 pb-3" style:min-height="50px">
  {#if hasFilters}
    <button
      transition:fly|local={{ duration: 200, y: 5 }}
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
