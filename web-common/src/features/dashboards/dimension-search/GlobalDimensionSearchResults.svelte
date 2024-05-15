<script lang="ts">
  import { Progress } from "@rilldata/web-common/components/progress";
  import DimensionSearchResult from "@rilldata/web-common/features/dashboards/dimension-search/DimensionSearchResult.svelte";
  import { useDimensionSearchResults } from "@rilldata/web-common/features/dashboards/dimension-search/useDimensionSearchResults";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import { toggleDimensionValueSelection } from "@rilldata/web-common/features/dashboards/state-managers/actions/dimension-filters";
  import type { DashboardMutables } from "@rilldata/web-common/features/dashboards/state-managers/actions/types";
  import { updateMetricsExplorerByName } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";

  export let metricsViewName: string;
  export let searchText: string;
  export let onSelect: () => void;

  $: instanceId = $runtime.instanceId;
  $: metricsViewQuery = useDashboard(instanceId, metricsViewName);

  let results: ReturnType<typeof useDimensionSearchResults>;
  $: if (
    $metricsViewQuery.data?.metricsView?.state?.validSpec &&
    !!searchText
  ) {
    results = useDimensionSearchResults(
      instanceId,
      metricsViewName,
      $metricsViewQuery.data.metricsView.state.validSpec,
      searchText,
    );
  }

  $: responses = $results?.responses.filter((r) => r?.values?.length) ?? [];

  function onItemSelect(dimension: string, value: any) {
    onSelect();
    updateMetricsExplorerByName(metricsViewName, (dashboard) => {
      toggleDimensionValueSelection(
        { dashboard } as DashboardMutables,
        dimension,
        value,
        false,
        true,
      );
    });
  }
</script>

<DropdownMenu open={!!$results}>
  <DropdownMenuTrigger />
  <DropdownMenuContent class="w-[450px] max-h-96 overflow-scroll">
    <div class="flex flex-col gap-2">
      {#if $results.completed && responses.length === 0}
        <div class="ui-copy-disabled text-center p-2 w-full">no results</div>
      {:else}
        {#if $results.progress < 100}
          <div class="flex flex-row items-center gap-x-2">
            <Progress value={$results.progress} max={100} />
            {$results.progress}%
          </div>
          <DropdownMenuSeparator />
        {/if}
        {#each responses as { dimension, values } (dimension)}
          <DimensionSearchResult {dimension} {values} onSelect={onItemSelect} />
          <DropdownMenuSeparator />
        {/each}
      {/if}
    </div>
  </DropdownMenuContent>
</DropdownMenu>
