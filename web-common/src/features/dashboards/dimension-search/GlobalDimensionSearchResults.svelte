<script lang="ts">
  import DimensionSearchResult from "@rilldata/web-common/features/dashboards/dimension-search/DimensionSearchResult.svelte";
  import { useDimensionSearchResults } from "@rilldata/web-common/features/dashboards/dimension-search/useDimensionSearchResults";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";

  export let metricsViewName: string;
  export let searchText: string;

  $: instanceId = $runtime.instanceId;
  $: dashboard = useDashboard(instanceId, metricsViewName);

  let results: ReturnType<typeof useDimensionSearchResults>;
  $: if ($dashboard.data?.metricsView?.state?.validSpec && !!searchText) {
    results = useDimensionSearchResults(
      instanceId,
      metricsViewName,
      $dashboard.data.metricsView.state.validSpec,
      searchText,
    );
  }

  $: responses = $results?.responses.filter((r) => r?.values?.length) ?? [];
</script>

<DropdownMenu open={!!$results}>
  <DropdownMenuTrigger />
  <DropdownMenuContent class="w-[450px] max-h-96 overflow-scroll">
    <div class="flex flex-col gap-2">
      {#if $results.progress < 100}
        {$results.progress}%
        <DropdownMenuSeparator />
      {/if}
      {#each responses as { dimension, values } (dimension)}
        <DimensionSearchResult {dimension} {values} />
        <DropdownMenuSeparator />
      {/each}
    </div>
  </DropdownMenuContent>
</DropdownMenu>
