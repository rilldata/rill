<script lang="ts">
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import {
    useQueryServiceColumnTimeRange,
    useRuntimeServiceGetCatalogEntry,
  } from "../../../runtime-client";
  import { metricsExplorerStore, useDashboardStore } from "../dashboard-stores";
  import { ComparisonRange, TimeRangeName } from "./time-control-types";
  import {
    exclusiveToInclusiveEndISOString,
    getComparisonTimeRange,
    getDateFromISOString,
  } from "./time-range-utils";
  import TimeRangeSelector from "./TimeRangeSelector.svelte";

  export let metricViewName;
  export let comparisonOptions: ComparisonRange[];

  $: dashboardStore = useDashboardStore(metricViewName);

  $: selectedTimeRange = $dashboardStore?.selectedTimeRange;

  let metricsViewQuery;
  $: if ($runtimeStore?.instanceId) {
    metricsViewQuery = useRuntimeServiceGetCatalogEntry(
      $runtimeStore.instanceId,
      metricViewName
    );
  }
  let timeRangeQuery;
  $: if (
    $runtimeStore?.instanceId &&
    $metricsViewQuery?.data?.entry?.metricsView?.model &&
    $metricsViewQuery?.data?.entry?.metricsView?.timeDimension
  ) {
    timeRangeQuery = useQueryServiceColumnTimeRange(
      $runtimeStore.instanceId,
      $metricsViewQuery.data.entry.metricsView.model,
      {
        columnName: $metricsViewQuery.data.entry.metricsView.timeDimension,
      }
    );
  }

  $: min = $timeRangeQuery.data.timeRangeSummary?.min
    ? getDateFromISOString($timeRangeQuery.data.timeRangeSummary.min)
    : undefined;

  $: options = comparisonOptions.map((comparisonRange) => {
    const comparisonTimeRange = getComparisonTimeRange(
      selectedTimeRange,
      comparisonRange
    );
    return {
      name: comparisonRange,
      start: comparisonTimeRange.start,
      end: comparisonTimeRange.end,
    };
  });

  let initialStartDate;
  let initialEndDate;
  $: if (selectedTimeRange) {
    initialStartDate = getDateFromISOString(
      $dashboardStore.selectedComparisonTimeRange.start
    );
    initialEndDate = getDateFromISOString(
      exclusiveToInclusiveEndISOString(
        $dashboardStore.selectedComparisonTimeRange.end
      )
    );
  }

  const onCompareRangeSelect = (timeRange) => {
    let name = timeRange.name;
    if (timeRange.name === TimeRangeName.Custom) {
      name = ComparisonRange.Custom;
    }

    metricsExplorerStore.setSelectedComparisonRange(metricViewName, {
      ...timeRange,
      name,
    });
  };

  // Define a better validation criteria
  function validateCustomTimeRange(start, end) {
    const customStartDate = new Date(start);
    const customEndDate = new Date(end);
    const selectedEndDate = new Date(selectedTimeRange.end);

    if (customStartDate > customEndDate)
      return "Start date must be before end date";
    else if (customEndDate > selectedEndDate)
      return "End date must be before selected date";
    else return undefined;
  }
</script>

<div class="flex gap-x-2 flex-row items-center pl-3">
  <div>Compare to</div>

  <TimeRangeSelector
    on:select-time-range={(e) => {
      onCompareRangeSelect(e.detail);
    }}
    timeRangeOptions={options}
    selectedTimeRange={$dashboardStore.selectedComparisonTimeRange}
    {min}
    max={selectedTimeRange.end}
    {initialStartDate}
    {initialEndDate}
    {validateCustomTimeRange}
  />
</div>
