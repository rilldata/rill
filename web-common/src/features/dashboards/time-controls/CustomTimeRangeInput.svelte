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
      start = getDateFromISOString(metricsExplorer.selectedTimeRange.start);
      end = getDateFromISOString(
        exclusiveToInclusiveEndISOString(metricsExplorer.selectedTimeRange.end)
      );
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

  $: min = $timeRangeQuery.data.timeRangeSummary?.min
    ? getDateFromISOString($timeRangeQuery.data.timeRangeSummary.min)
    : undefined;
  $: max = $timeRangeQuery.data.timeRangeSummary?.max
    ? getDateFromISOString($timeRangeQuery.data.timeRangeSummary.max)
    : undefined;

  function applyCustomTimeRange() {
    // Currently, we assume UTC
    dispatch("apply", {
      startDate: getISOStringFromDate(start),
      endDate: getISOStringFromDate(end),
    });
  }

  function exclusiveToInclusiveEndISOString(exclusiveEnd: string): string {
    const date = new Date(exclusiveEnd);
    date.setDate(date.getDate() - 1);
    return date.toISOString();
  }

  function getDateFromISOString(isoString: string): string {
    return isoString.split("T")[0];
  }

  function getISOStringFromDate(date: string): string {
    return date + "T00:00:00.000Z";
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
      on:blur={() => dispatch("close-calendar")}
      type="date"
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
      on:blur={() => dispatch("close-calendar")}
      type="date"
      id="end-date"
      name="end-date"
      {min}
      {max}
    />
  </div>
  <div class="flex mt-1">
    <div class="flex-grow" />
    <Button type="primary" submitForm form="custom-time-range-form" {disabled}>
      Apply
    </Button>
  </div>
</form>
