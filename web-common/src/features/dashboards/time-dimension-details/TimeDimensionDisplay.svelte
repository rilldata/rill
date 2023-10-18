<script lang="ts">
  import {
    metricsExplorerStore,
    useDashboardStore,
  } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
  import TDDHeader from "./TDDHeader.svelte";
  import TddNew from "./TDDNew.svelte";
  import { getStateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
  import {
    tableInteractionStore,
    useTimeDimensionDataStore,
  } from "./time-dimension-data-store";
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

  // Get labels for table headers
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

  // Create a copy of the data to avoid flashing of table in transient states
  let timeDimensionDataCopy;
  $: if ($timeDimensionDataStore?.data) {
    timeDimensionDataCopy = $timeDimensionDataStore.data;
  }
  $: formattedData = timeDimensionDataCopy;

  $: excludeMode =
    $dashboardStore?.dimensionFilterExcludeMode.get(dimensionName) ?? false;

  // Transform the scrub dates into table position
  $: columnHeaders = formattedData?.columnHeaderData?.flat();
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

  // Create a time formatter for the column headers
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

    const dimensionValue = formattedData?.rowHeaderData[y]?.[0]?.value;
    const time =
      x && columnHeaders[x]?.value && new Date(columnHeaders[x]?.value);

    tableInteractionStore.set({
      dimensionValue,
      time,
    });
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
