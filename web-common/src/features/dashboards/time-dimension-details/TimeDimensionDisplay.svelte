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
  import { useMetaQuery } from "@rilldata/web-common/features/dashboards/selectors/index";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";

  export let metricViewName;

  const timeDimensionDataStore = useTimeDimensionDataStore(getStateManagers());
  $: metaQuery = useMetaQuery(getStateManagers());
  $: dashboardStore = useDashboardStore(metricViewName);
  $: dimensionName = $dashboardStore?.selectedComparisonDimension;

  $: measureLabel =
    $metaQuery?.data?.measures?.find(
      (m) => m.name === $dashboardStore?.expandedMeasureName
    ).label || "";

  let dimensionLabel = "";
  $: if ($timeDimensionDataStore?.comparing === "dimension") {
    dimensionLabel = $metaQuery?.data?.dimensions?.find(
      (d) => d.name === dimensionName
    ).label;
  } else if ($timeDimensionDataStore?.comparing === "time") {
    dimensionLabel = "Time";
  } else {
    dimensionLabel = "No Comparison";
  }
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
</script>

<TDDHeader
  {dimensionName}
  {metricViewName}
  comparing={$timeDimensionDataStore?.comparing}
  on:search={(e) => {
    metricsExplorerStore.setSearchText(metricViewName, e.detail);
  }}
/>

{#if $timeDimensionDataStore?.data}
  <TddNew
    on:toggle-sort={() =>
      metricsExplorerStore.toggleSort(metricViewName, SortType.VALUE)}
    {dimensionName}
    {metricViewName}
    {excludeMode}
    {dimensionLabel}
    {measureLabel}
    scrubPos={{ start: startScrubPos, end: endScrubPos }}
    sortDirection={$dashboardStore.sortDirection === SortDirection.ASCENDING}
    comparing={$timeDimensionDataStore?.comparing}
    timeFormatter={$timeDimensionDataStore.timeFormatter}
    data={$timeDimensionDataStore.data}
  />
{:else}
  <Spinner size="18px" status={EntityStatus.Running} />
{/if}
