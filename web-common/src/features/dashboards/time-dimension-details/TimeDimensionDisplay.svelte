<script lang="ts">
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import TDDHeader from "./TDDHeader.svelte";
  import TddNew from "./TDDNew.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import { useTimeDimensionDataStore } from "./time-dimension-data-store";
  import {
    SortDirection,
    SortType,
  } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
  import { bisectData } from "@rilldata/web-common/components/data-graphic/utils";

  export let metricViewName;

  const timeDimensionDataStore = useTimeDimensionDataStore(getStateManagers());
  $: dashboardStore = useDashboardStore(metricViewName);
  $: dimensionName = $dashboardStore?.selectedComparisonDimension;

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  $: startScrubPos = bisectData(
    $dashboardStore?.selectedScrubRange?.start,
    "center",
    "value",
    $timeDimensionDataStore?.data?.columnHeaderData?.flat(),
    true
  );

  $: endScrubPos = bisectData(
    $dashboardStore?.selectedScrubRange?.end,
    "center",
    "value",
    $timeDimensionDataStore?.data?.columnHeaderData?.flat(),
    true
  );

  $: console.log(startScrubPos, endScrubPos);
</script>

<TDDHeader {dimensionName} {metricViewName} />

{#if $timeDimensionDataStore?.data}
  <TddNew
    on:toggle-sort={() =>
      metricsExplorerStore.toggleSort(metricViewName, SortType.VALUE)}
    {dimensionName}
    {metricViewName}
    {excludeMode}
    scrubPos={{ start: startScrubPos, end: endScrubPos }}
    sortDirection={$dashboardStore.sortDirection === SortDirection.ASCENDING}
    comparing={$timeDimensionDataStore.comparing}
    timeFormatter={$timeDimensionDataStore.timeFormatter}
    data={$timeDimensionDataStore.data}
  />
{/if}
