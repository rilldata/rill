<script lang="ts">
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetTimeRangeSummary,
    V1GetTimeRangeSummaryResponse,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useDashboardStore } from "../dashboard-stores";
  import {
    exclusiveToInclusiveEndISOString,
    getDateFromISOString,
    getISOStringFromDate,
  } from "./time-range-utils";

  export let metricViewName: string;

  const dispatch = createEventDispatcher();

  let start: string;
  let end: string;

  $: dashboardStore = useDashboardStore(metricViewName);

  $: if (!start && !end) {
    if ($dashboardStore?.selectedTimeRange) {
      start = getDateFromISOString($dashboardStore.selectedTimeRange.start);
      end = getDateFromISOString(
        exclusiveToInclusiveEndISOString($dashboardStore.selectedTimeRange.end)
      );
    }
  }

  $: error = validateTimeRange(start, end);
  $: disabled = !start || !end || !!error;

  let metricsViewQuery;
  $: if ($runtime?.instanceId) {
    metricsViewQuery = useRuntimeServiceGetCatalogEntry(
      $runtime.instanceId,
      metricViewName
    );
  }
  let timeRangeQuery: UseQueryStoreResult<V1GetTimeRangeSummaryResponse, Error>;
  $: if (
    $runtime?.instanceId &&
    $metricsViewQuery?.data?.entry?.metricsView?.model &&
    $metricsViewQuery?.data?.entry?.metricsView?.timeDimension
  ) {
    timeRangeQuery = useRuntimeServiceGetTimeRangeSummary(
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

  function validateTimeRange(start: string, end: string) {
    if (start > end) {
      return "Start date must be before end date";
    } else {
      return undefined;
    }
  }

  function applyCustomTimeRange() {
    // Currently, we assume UTC
    dispatch("apply", {
      startDate: getISOStringFromDate(start),
      endDate: getISOStringFromDate(end),
    });
  }

  let labelClasses = "font-semibold text-[10px]";
</script>

<form
  id="custom-time-range-form"
  class="flex flex-col gap-y-3 mt-3 mb-1 px-3"
  on:submit|preventDefault={applyCustomTimeRange}
>
  <div class="flex flex-col gap-y-1">
    <label for="start-date" class={labelClasses}>Start date</label>
    <input
      bind:value={start}
      on:blur={() => dispatch("close-calendar")}
      type="date"
      id="start-date"
      name="start-date"
      {min}
      {max}
      class="cursor-pointer"
    />
  </div>
  <div class="flex flex-col gap-y-1">
    <label for="end-date" class={labelClasses}>End date</label>

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
  <div class="flex mt-3 items-center">
    {#if error}
      <div style:font-size="11px" class="text-red-600">
        {error}
      </div>
    {/if}
    <div class="flex-grow" />
    <Button type="primary" submitForm form="custom-time-range-form" {disabled}>
      Apply
    </Button>
  </div>
</form>
