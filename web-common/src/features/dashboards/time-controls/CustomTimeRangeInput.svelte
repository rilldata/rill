<script lang="ts">
  import type { UseQueryStoreResult } from "@sveltestack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import {
    useQueryServiceColumnTimeRange,
    useRuntimeServiceGetCatalogEntry,
    V1ColumnTimeRangeResponse,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useDashboardStore } from "../dashboard-stores";
  import {
    exclusiveToInclusiveEndISOString,
    getDateFromISOString,
    getISOStringFromDate,
    validateTimeRange,
  } from "./time-range-utils";

  export let metricViewName: string;
  export let minTimeGrain: string;

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

  $: error = validateTimeRange(new Date(start), new Date(end), minTimeGrain);
  $: disabled = !start || !end || !!error;

  let metricsViewQuery;
  $: if ($runtime?.instanceId) {
    metricsViewQuery = useRuntimeServiceGetCatalogEntry(
      $runtime.instanceId,
      metricViewName
    );
  }
  let timeRangeQuery: UseQueryStoreResult<V1ColumnTimeRangeResponse, Error>;
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
  class="flex flex-col gap-y-3 mt-3 mb-1 px-3"
  id="custom-time-range-form"
  on:submit|preventDefault={applyCustomTimeRange}
>
  <div class="flex flex-col gap-y-1">
    <label class={labelClasses} for="start-date">Start date</label>
    <input
      bind:value={start}
      class="cursor-pointer"
      id="start-date"
      {max}
      {min}
      name="start-date"
      on:blur={() => dispatch("close-calendar")}
      type="date"
    />
  </div>
  <div class="flex flex-col gap-y-1">
    <label class={labelClasses} for="end-date">End date</label>

    <input
      bind:value={end}
      id="end-date"
      {max}
      {min}
      name="end-date"
      on:blur={() => dispatch("close-calendar")}
      type="date"
    />
  </div>
  <div class="flex mt-3 items-center">
    {#if error}
      <div style:font-size="11px" class="text-red-600">
        {error}
      </div>
    {/if}
    <div class="flex-grow" />
    <Button {disabled} form="custom-time-range-form" submitForm type="primary">
      Apply
    </Button>
  </div>
</form>
