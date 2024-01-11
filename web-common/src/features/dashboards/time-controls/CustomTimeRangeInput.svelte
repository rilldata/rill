<script lang="ts">
  import {
    getAllowedTimeGrains,
    isGrainBigger,
  } from "@rilldata/web-common/lib/time/grains";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import {
    DashboardTimeControls,
    Period,
    TimeOffsetType,
  } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "../../../runtime-client";
  import { parseLocaleStringDate } from "@rilldata/web-common/components/date-picker/util";
  import { getOffset } from "@rilldata/web-common/lib/time/transforms";
  import { removeZoneOffset } from "@rilldata/web-common/lib/time/timezone";

  export let minTimeGrain: V1TimeGrain;
  export let boundaryStart: Date;
  export let boundaryEnd: Date;
  export let defaultDate: DashboardTimeControls;
  export let zone: string;

  const dispatch = createEventDispatcher();

  let start: string;
  let end: string;

  let error: string | undefined = undefined;

  $: disabled = !start || !end || !!error;

  $: max = getDateFromISOString(boundaryEnd.toISOString());
  $: min = getDateFromISOString(boundaryStart.toISOString());

  $: if (!start && !end && defaultDate) {
    start = getDateFromObject(defaultDate.start);
    end = getDateFromObject(defaultDate.end, true);
  }

  $: if (start && end) {
    error = validateTimeRange(
      parseLocaleStringDate(start),
      getOffset(
        new Date(getISOStringFromDate(end, "UTC")),
        Period.DAY,
        TimeOffsetType.ADD,
      ),
      minTimeGrain,
    );
  }

  export function getDateFromObject(date: Date, exclusive = false): string {
    if (exclusive) {
      date = new Date(date.getTime() - 1);
    }

    return getDateFromISOString(date.toISOString());
  }

  export function getDateFromISOString(isoDate: string): string {
    return isoDate.split("T")[0];
  }

  export function getISOStringFromDate(
    date: string,
    timeZone?: string,
  ): string {
    return parseLocaleStringDate(date, timeZone).toISOString();
  }

  function validateTimeRange(
    start: Date,
    end: Date,
    minTimeGrain: V1TimeGrain,
  ): string | undefined {
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

  function applyCustomTimeRange() {
    let startDate = getISOStringFromDate(start, "UTC");
    let endDate = getOffset(
      new Date(getISOStringFromDate(end, "UTC")),
      Period.DAY,
      TimeOffsetType.ADD,
    ).toISOString();

    startDate = removeZoneOffset(new Date(startDate), zone).toISOString();
    endDate = removeZoneOffset(new Date(endDate), zone).toISOString();

    dispatch("apply", {
      startDate,
      endDate,
    });
  }
</script>

<form
  id="custom-time-range-form"
  class="relative mb-1 mt-3 flex flex-col gap-y-3 px-3"
  on:submit|preventDefault={applyCustomTimeRange}
>
  <div class="flex flex-row gap-x-3">
    <div class="relative flex flex-col gap-y-1">
      <label class="text-[10px] font-semibold" for="start-date">
        Start date
      </label>
      <input
        type="date"
        {max}
        {min}
        id="start-date"
        name="start-date"
        bind:value={start}
        on:blur={() => dispatch("close-calendar")}
      />
    </div>
    <div class="relative flex flex-col gap-y-1">
      <label class="text-[10px] font-semibold" for="start-date">
        End date
      </label>
      <input
        type="date"
        {max}
        {min}
        id="start-date"
        name="start-date"
        bind:value={end}
        on:blur={() => dispatch("close-calendar")}
      />
    </div>
  </div>

  <div class="mt-3 flex items-center">
    {#if error}
      <div style:font-size="11px" class="mr-2 text-red-600">
        {error}
      </div>
    {/if}
    <div class="flex-grow" />

    <Button {disabled} form="custom-time-range-form" submitForm type="primary">
      Apply
    </Button>
  </div>
</form>
