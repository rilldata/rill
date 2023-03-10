<script lang="ts">
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

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
  $: if ($dashboardStore?.selectedTimeRange?.start) {
    initialStartDate = getDateFromISOString(
      $dashboardStore.selectedTimeRange.start
    );
    initialEndDate = getDateFromISOString(
      exclusiveToInclusiveEndISOString($dashboardStore.selectedTimeRange.end)
    );
  }

  let metricsViewQuery;
  $: if ($runtime?.instanceId) {
    metricsViewQuery = useRuntimeServiceGetCatalogEntry(
      $runtime.instanceId,
      metricViewName
    );
  }
  let timeRangeQuery;
  $: if (
    $runtime?.instanceId &&
    $metricsViewQuery?.data?.entry?.metricsView?.model &&
    $metricsViewQuery?.data?.entry?.metricsView?.timeDimension
  ) {
    timeRangeQuery = useQueryServiceColumnTimeRange(
      $runtime.instanceId,
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
  {initialEndDate}
  {initialStartDate}
  {max}
  {min}
  on:select-time-range
  {selectedTimeRange}
  showIcon
  showPreciseRange
  timeRangeOptions={relativeTimeRangeOptions}
  {validateCustomTimeRange}
/>
