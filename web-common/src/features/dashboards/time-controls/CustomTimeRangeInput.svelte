<script lang="ts">
  import {
    getAllowedTimeGrains,
    isGrainBigger,
  } from "@rilldata/web-common/lib/time/grains";
  import { getOffset } from "@rilldata/web-common/lib/time/transforms";
  import { TimeOffsetType } from "@rilldata/web-common/lib/time/types";
  import type { CreateQueryResult } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import {
    createQueryServiceColumnTimeRange,
    createRuntimeServiceGetCatalogEntry,
    V1ColumnTimeRangeResponse,
    V1TimeGrain,
  } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useDashboardStore } from "../dashboard-stores";

  export let metricViewName: string;
  export let minTimeGrain: V1TimeGrain;

  const dispatch = createEventDispatcher();

  let start: string;
  let end: string;

  $: dashboardStore = useDashboardStore(metricViewName);

  $: if (!start && !end && $dashboardStore?.selectedTimeRange.start) {
    start = getDateFromObject($dashboardStore.selectedTimeRange.start);
    end = getDateFromObject(
      getOffset(
        new Date($dashboardStore.selectedTimeRange.end),
        "P1D",
        TimeOffsetType.SUBTRACT
      )
    );
  }

  // functions for extracting the right kind of date string out of
  // a Date object. Used in the input elements.
  export function getDateFromObject(date: Date): string {
    return getDateFromISOString(date.toISOString());
  }

  export function getDateFromISOString(isoDate: string): string {
    return isoDate.split("T")[0];
  }

  export function getISOStringFromDate(date: string): string {
    return date + "T00:00:00.000Z";
  }

  function validateTimeRange(
    start: Date,
    end: Date,
    minTimeGrain: V1TimeGrain
  ): string {
    const allowedTimeGrains = getAllowedTimeGrains(start, end);
    const allowedMaxGrain = allowedTimeGrains[allowedTimeGrains.length - 1];

    const isGrainPossible = !isGrainBigger(minTimeGrain, allowedMaxGrain.grain);

    if (start > end) {
      return "Start date must be before end date";
    } else if (!isGrainPossible) {
      return "Range is smaller than min time grain";
    } else {
      return undefined;
    }
  }

  // HAM, you left off here.
  $: error = validateTimeRange(new Date(start), new Date(end), minTimeGrain);
  $: disabled = !start || !end || !!error;

  let metricsViewQuery;
  $: if ($runtime?.instanceId) {
    metricsViewQuery = createRuntimeServiceGetCatalogEntry(
      $runtime.instanceId,
      metricViewName
    );
  }
  let timeRangeQuery: CreateQueryResult<V1ColumnTimeRangeResponse, Error>;
  $: if (
    $runtime?.instanceId &&
    $metricsViewQuery?.data?.entry?.metricsView?.model &&
    $metricsViewQuery?.data?.entry?.metricsView?.timeDimension
  ) {
    timeRangeQuery = createQueryServiceColumnTimeRange(
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
    const startDate = getISOStringFromDate(start);
    const endDate = getISOStringFromDate(end);
    dispatch("apply", {
      startDate,
      endDate,
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
