<script lang="ts">
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import { createEventDispatcher } from "svelte";

  import {
    useQueryServiceColumnTimeRange,
    useRuntimeServiceGetCatalogEntry,
    V1TimeGrain,
  } from "../../../runtime-client";
  import { useDashboardStore } from "../dashboard-stores";

  import type { TimeRange } from "./time-control-types";
  import {
    exclusiveToInclusiveEndISOString,
    getDateFromISOString,
    getRelativeTimeRangeOptions,
    validateTimeRange,
  } from "./time-range-utils";
  import TimeRangeSelector from "./TimeRangeSelector.svelte";

  export let metricViewName: string;
  export let allTimeRange: TimeRange;
  export let minTimeGrain: V1TimeGrain;

  $: dashboardStore = useDashboardStore(metricViewName);

  $: selectedTimeRange = $dashboardStore?.selectedTimeRange;
  let relativeTimeRangeOptions: TimeRange[];
  $: if (allTimeRange) {
    relativeTimeRangeOptions = getRelativeTimeRangeOptions(
      allTimeRange,
      minTimeGrain
    );
  }

  let initialStartDate;
  let initialEndDate;
  $: if ($dashboardStore?.selectedTimeRange) {
    initialStartDate = getDateFromISOString(
      $dashboardStore.selectedTimeRange.start
    );
    initialEndDate = getDateFromISOString(
      exclusiveToInclusiveEndISOString($dashboardStore.selectedTimeRange.end)
    );
  }

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
  $: max = $timeRangeQuery.data.timeRangeSummary?.max
    ? getDateFromISOString($timeRangeQuery.data.timeRangeSummary.max)
    : undefined;

  function validateCustomTimeRange(start: string, end: string) {
    return validateTimeRange(new Date(start), new Date(end), minTimeGrain);
  }
</script>

<TimeRangeSelector
  on:select-time-range
  timeRangeOptions={relativeTimeRangeOptions}
  {selectedTimeRange}
  {min}
  {max}
  {initialStartDate}
  {initialEndDate}
  {validateCustomTimeRange}
/>
