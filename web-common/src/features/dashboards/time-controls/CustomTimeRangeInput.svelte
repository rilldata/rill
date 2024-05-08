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
  import Litepicker from "@rilldata/web-common/components/date-picker/Litepicker.svelte";
  import {
    parseLocaleStringDate,
    shiftToUTC,
  } from "@rilldata/web-common/components/date-picker/util";
  import { getOffset } from "@rilldata/web-common/lib/time/transforms";
  import { removeZoneOffset } from "@rilldata/web-common/lib/time/timezone";

  export let minTimeGrain: V1TimeGrain;
  export let boundaryStart: Date;
  export let boundaryEnd: Date;
  export let defaultDate: DashboardTimeControls | undefined;
  export let zone: string;

  const dispatch = createEventDispatcher();

  let start: string;
  let end: string;

  $: if (!start && !end && defaultDate) {
    start = getDateFromObject(defaultDate.start);
    end = getDateFromObject(defaultDate.end, true);
  }

  // functions for extracting the right kind of date string out of
  // a Date object. Used in the input elements.
  export function getDateFromObject(date: Date, exclusive = false): string {
    if (exclusive) {
      date = new Date(date.getTime() - 1);
    }
    return date.toLocaleDateString(
      Intl.DateTimeFormat().resolvedOptions().locale,
      {
        timeZone: "UTC",
      },
    );
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

  // HAM, you left off here.
  let error: string | undefined = undefined;
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

  $: disabled = !start || !end || !!error;

  $: max = getDateFromISOString(boundaryEnd.toISOString());
  $: min = getDateFromISOString(boundaryStart.toISOString());

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

  let startEl, endEl, editingDate, isOpen;

  const handleDatePickerChange = (d) => {
    start = getDateFromObject(shiftToUTC(d.detail.start));
    end = getDateFromObject(shiftToUTC(d.detail.end));
  };

  const handleEditingChange = (d) => {
    editingDate = d.detail;
  };

  const handleToggle = (d) => {
    isOpen = d.detail;
  };

  let labelClasses = "font-semibold text-[10px]";
  $: getInputClasses = (v) =>
    `cursor-pointer w-full ${
      isOpen && v === editingDate ? "input-outline" : ""
    } `;

  const handleInputKeyDown = (e) => {
    if (e.key === "Enter") {
      e.currentTarget.blur();
    }
  };
</script>

<form
  class="flex flex-col gap-y-3 mt-3 mb-1 px-3 relative"
  id="custom-time-range-form"
  on:submit|preventDefault={applyCustomTimeRange}
>
  <div class="flex flex-row gap-x-3">
    <div class="flex flex-col gap-y-1 relative">
      <label class={labelClasses} for="start-date">Start date</label>
      <input
        bind:this={startEl}
        class={getInputClasses(0)}
        id="start-date"
        {max}
        {min}
        name="start-date"
        on:blur={() => dispatch("close-calendar")}
        on:keydown={handleInputKeyDown}
        type="text"
      />
    </div>

    <div class="flex flex-col gap-y-1 relative">
      <label class={labelClasses} for="end-date">End date</label>
      <input
        bind:this={endEl}
        id="end-date"
        {min}
        {max}
        name="end-date"
        class={getInputClasses(1)}
        on:blur={() => dispatch("close-calendar")}
        on:keydown={handleInputKeyDown}
        type="text"
      />
    </div>
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
    <Litepicker
      {startEl}
      {endEl}
      min={getDateFromObject(boundaryStart)}
      max={getDateFromObject(boundaryEnd)}
      defaultStart={start}
      defaultEnd={end}
      openOnMount
      on:change={handleDatePickerChange}
      on:editing={handleEditingChange}
      on:toggle={handleToggle}
    />
  {/if}
</form>

<style>
  .input-outline {
    outline-offset: 2px;
    /* FF */
    outline: Highlight auto 1px;
    /* Chrome/Safari */
    outline: -webkit-focus-ring-color auto 1px;
  }
</style>
