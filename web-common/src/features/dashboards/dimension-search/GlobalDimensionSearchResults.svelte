<script lang="ts">
  import { Progress } from "@rilldata/web-common/components/progress";
  import GlobalDimensionSearchResult from "@rilldata/web-common/features/dashboards/dimension-search/GlobalDimensionSearchResult.svelte";
  import {
    DimensionSearchResult,
    useDimensionSearchResults,
  } from "@rilldata/web-common/features/dashboards/dimension-search/useDimensionSearchResults";
  import { useDashboard } from "@rilldata/web-common/features/dashboards/selectors";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuTrigger,
  } from "@rilldata/web-common/components/dropdown-menu";

  export let metricsViewName: string;
  export let searchText: string;
  export let onSelect: () => void;
  export let open: boolean;

  const {
    actions: {
      dimensionsFilter: { toggleDimensionValueSelection },
    },
    timeRangeSummaryStore,
  } = getStateManagers();

  $: instanceId = $runtime.instanceId;
  $: metricsViewQuery = useDashboard(instanceId, metricsViewName);

  let results: ReturnType<typeof useDimensionSearchResults>;
  $: if (
    $metricsViewQuery.data?.metricsView?.state?.validSpec &&
    $timeRangeSummaryStore.data?.timeRangeSummary &&
    !!searchText
  ) {
    results = useDimensionSearchResults(
      instanceId,
      metricsViewName,
      $metricsViewQuery.data.metricsView.state.validSpec,
      $timeRangeSummaryStore.data.timeRangeSummary,
      searchText,
    );
  }

  $: responses = ($results?.responses.filter((r) => r?.values?.length) ??
    []) as DimensionSearchResult[];

  function onItemSelect(dimension: string, value: any) {
    onSelect();
    toggleDimensionValueSelection(dimension, value, false, true);
  }
</script>

<DropdownMenu
  open={open && !!$results && !!searchText}
  onOpenChange={(o) => (open = o)}
>
  <DropdownMenuTrigger asChild let:builder>
    <button use:builder.action {...builder} class="absolute left-32"></button>
  </DropdownMenuTrigger>
  <DropdownMenuContent
    class="w-64 max-h-96 overflow-scroll right-2"
    sideOffset={32}
  >
    <div class="flex flex-col divide-y divide-slate-200">
      {#if $results.errors.length}
        <div class="text-center p-2 w-full text-red-500">
          Search error. Try again.
        </div>
      {:else if $results.completed && responses.length === 0}
        <div class="ui-copy-disabled text-center p-2 w-full">no results</div>
      {:else}
        {#if $results.progress < 100}
          <div class="flex flex-row items-center gap-x-2 px-2">
            <Progress value={$results.progress} max={100} class="h-1" />
            <div class="text-gray-500 text-[11px]">{$results.progress}%</div>
          </div>
        {/if}
        {#each responses as { dimension, values } (dimension)}
          <GlobalDimensionSearchResult
            {dimension}
            {values}
            onSelect={onItemSelect}
          />
        {/each}
      {/if}
    </div>
  </DropdownMenuContent>
</DropdownMenu>
