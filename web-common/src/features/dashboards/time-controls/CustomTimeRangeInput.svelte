<script lang="ts">
  import { runtimeStore } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetTimeRangeSummary,
    V1GetTimeRangeSummaryResponse,
  } from "../../../runtime-client";
  import { metricsExplorerStore } from "../dashboard-stores";

  export let metricViewName: string;

  const dispatch = createEventDispatcher();

  let start: string;
  let end: string;

  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];
  $: if (!start && !end) {
    if (metricsExplorer?.selectedTimeRange) {
      start = stripUTCTimezone(metricsExplorer.selectedTimeRange.start);
      end = stripUTCTimezone(metricsExplorer.selectedTimeRange.end);
    }
  }

  $: disabled = !start || !end;

  let metricsViewQuery;
  $: if ($runtimeStore?.instanceId) {
    metricsViewQuery = useRuntimeServiceGetCatalogEntry(
      $runtimeStore.instanceId,
      metricViewName
    );
  }
  let timeRangeQuery: UseQueryStoreResult<V1GetTimeRangeSummaryResponse, Error>;
  $: if (
    $runtimeStore?.instanceId &&
    $metricsViewQuery?.data?.entry?.metricsView?.model &&
    $metricsViewQuery?.data?.entry?.metricsView?.timeDimension
  ) {
    timeRangeQuery = useRuntimeServiceGetTimeRangeSummary(
      $runtimeStore.instanceId,
      $metricsViewQuery.data.entry.metricsView.model,
      {
        columnName: $metricsViewQuery.data.entry.metricsView.timeDimension,
      }
    );
  }

  // <input type=datetime-local/> does not accept time zone information
  $: min = $timeRangeQuery.data.timeRangeSummary?.min
    ? stripUTCTimezone($timeRangeQuery.data.timeRangeSummary.min)
    : undefined;
  $: max = $timeRangeQuery.data.timeRangeSummary?.min
    ? stripUTCTimezone($timeRangeQuery.data.timeRangeSummary.max)
    : undefined;

  function applyCustomTimeRange() {
    // Currently, we assume UTC
    dispatch("apply", {
      startDate: addUTCTimezone(start),
      endDate: addUTCTimezone(end),
    });
  }

  function stripUTCTimezone(date: string) {
    return date.replace(/Z$/, "");
  }

  function addUTCTimezone(date: string) {
    return date + "Z";
  }
</script>

<form
  id="custom-time-range-form"
  class="flex flex-col gap-y-3 mt-3 mb-1 px-3"
  on:submit|preventDefault={applyCustomTimeRange}
>
  <div class="flex flex-col gap-y-1">
    <label for="start-date" style="font-size: 10px;">Start date</label>
    <input
      bind:value={start}
      type="datetime-local"
      id="start-date"
      name="start-date"
      {min}
      {max}
    />
  </div>
  <div class="flex flex-col gap-y-1">
    <label for="end-date" style="font-size: 10px;">End date</label>
    <input
      bind:value={end}
      type="datetime-local"
      id="end-date"
      name="end-date"
      {min}
      {max}
    />
  </div>
  <div class="flex">
    <div class="flex-grow" />
    <Button type="primary" submitForm form="custom-time-range-form" {disabled}>
      Apply
    </Button>
  </div>
</form>
