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
  import {
    bisectData,
    createTimeFormat,
  } from "@rilldata/web-common/components/data-graphic/utils";
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

  $: columnHeaders = $timeDimensionDataStore?.data?.columnHeaderData?.flat();
  $: startScrubPos = bisectData(
    $dashboardStore?.selectedScrubRange?.start,
    "center",
    "value",
    columnHeaders,
    true
  );

  $: endScrubPos = bisectData(
    $dashboardStore?.selectedScrubRange?.end,
    "center",
    "value",
    columnHeaders,
    true
  );

  let timeDimensionDataCopy;
  $: if ($timeDimensionDataStore?.data) {
    timeDimensionDataCopy = $timeDimensionDataStore.data;
  }
  $: formattedData = timeDimensionDataCopy;

  $: timeFormatter =
    columnHeaders?.length &&
    createTimeFormat(
      [
        new Date(columnHeaders[0]?.value),
        new Date(columnHeaders[columnHeaders?.length - 1]?.value),
      ],
      columnHeaders?.length
    )[0];

  function highlightCell(e) {
    const { x, y } = e.detail;
  }
</script>

<TDDHeader
  {dimensionName}
  {metricViewName}
  comparing={$timeDimensionDataStore?.comparing}
  on:search={(e) => {
    metricsExplorerStore.setSearchText(metricViewName, e.detail);
  }}
/>

{#if formattedData}
  <TddNew
    {dimensionName}
    {metricViewName}
    {excludeMode}
    {dimensionLabel}
    {measureLabel}
    scrubPos={{ start: startScrubPos, end: endScrubPos }}
    sortDirection={$dashboardStore.sortDirection === SortDirection.ASCENDING}
    comparing={$timeDimensionDataStore?.comparing}
    {timeFormatter}
    data={formattedData}
    on:toggle-sort={() =>
      metricsExplorerStore.toggleSort(metricViewName, SortType.VALUE)}
    on:highlight={highlightCell}
  />
{:else}
  <Spinner size="18px" status={EntityStatus.Running} />
{/if}
