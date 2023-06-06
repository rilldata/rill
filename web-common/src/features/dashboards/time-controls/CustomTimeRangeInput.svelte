<script lang="ts">
  import {
    getAllowedTimeGrains,
    isGrainBigger,
  } from "@rilldata/web-common/lib/time/grains";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import type { DashboardTimeControls } from "../../../lib/time/types";
  import type { V1TimeGrain } from "../../../runtime-client";
  import DatePicker from "@rilldata/web-common/components/date-picker/DatePicker.svelte";

  export let minTimeGrain: V1TimeGrain;
  export let boundaryStart: Date;
  export let boundaryEnd: Date;
  export let defaultDate: DashboardTimeControls;

  const dispatch = createEventDispatcher();

  let start: string;
  let end: string;

  $: if (!start && !end && defaultDate) {
    start = getDateFromObject(defaultDate.start);
    end = getDateFromObject(defaultDate.end);
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
      return "Range is not valid for given min time grain";
    } else {
      return undefined;
    }
  }

  // HAM, you left off here.
  $: error = validateTimeRange(new Date(start), new Date(end), minTimeGrain);
  $: disabled = !start || !end || !!error;

  $: max = getDateFromISOString(boundaryEnd.toISOString());
  $: min = getDateFromISOString(boundaryStart.toISOString());

  function applyCustomTimeRange() {
    const startDate = getISOStringFromDate(start);
    const endDate = getISOStringFromDate(end);
    dispatch("apply", {
      startDate,
      endDate,
    });
  }

  let labelClasses = "font-semibold text-[10px]";

  let startEl, endEl;

  $: {
    console.log({ startEl, endEl });
  }
</script>

<form
  class="flex flex-col gap-y-3 mt-3 mb-1 px-3"
  id="custom-time-range-form"
  on:submit|preventDefault={applyCustomTimeRange}
>
  <div class="flex flex-col gap-y-1">
    <label class={labelClasses} for="start-date">Start date</label>
    <input
      bind:this={startEl}
      bind:value={start}
      class="cursor-pointer"
      id="start-date"
      {max}
      {min}
      name="start-date"
      on:blur={() => dispatch("close-calendar")}
      type="text"
    />
  </div>

  <div class="flex flex-col gap-y-1">
    <label class={labelClasses} for="end-date">End date</label>
    <input
      bind:this={endEl}
      bind:value={end}
      id="end-date"
      {min}
      {max}
      name="end-date"
      on:blur={() => dispatch("close-calendar")}
      type="text"
    />
  </div>
  <div class="flex mt-3 items-center">
    {#if error}
      <div style:font-size="11px" class="text-red-600 mr-2">
        {error}
      </div>
    {/if}
    <div class="flex-grow" />
    <Button {disabled} form="custom-time-range-form" submitForm type="primary">
      Apply
    </Button>
  </div>
  {#if startEl && endEl}
    <DatePicker {startEl} {endEl} />
  {/if}
</form>
