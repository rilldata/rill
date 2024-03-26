<script lang="ts">
  import {
    getAllowedTimeGrains,
    isGrainBigger,
  } from "@rilldata/web-common/lib/time/grains";
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeGrain } from "../../../runtime-client";
  import DateSelector from "@rilldata/web-common/components/date-picker/DateSelector.svelte";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Calendar from "@rilldata/web-common/components/date-picker/Calendar.svelte";
  import { DateTime } from "luxon";

  import { ChevronLeft, ChevronRight } from "lucide-svelte";

  const dispatch = createEventDispatcher();

  export let minTimeGrain: V1TimeGrain;
  export let boundaryStart: Date;
  export let boundaryEnd: Date;
  export let defaultDate: DashboardTimeControls;
  export let zone: string;

  let selecting: "start" | "end" = "start";
  let isCustomRangeOpen = false;
  let error: string | undefined = undefined;
  let start = DateTime.fromJSDate(defaultDate.start).setZone(zone);
  let end = DateTime.fromJSDate(defaultDate.end).setZone(zone);
  let nudgeGrain: "day" | "week" | "month" | "year" = "day";

  $: disabled = !start || !end || error !== undefined;

  $: if (start && end) {
    error = validateTimeRange(start.toJSDate(), end.toJSDate(), minTimeGrain);
  }

  function handleSubmit() {
    if (!end || !start) return;
    // Set the time to midnight to match existing behavior
    dispatch("apply", {
      startDate: start.set({ hour: 0, minute: 0, second: 0 }).toJSDate(),
      endDate: end.set({ hour: 0, minute: 0, second: 0 }).toJSDate(),
    });
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
</script>

<DropdownMenu.Sub bind:open={isCustomRangeOpen}>
  <DropdownMenu.SubTrigger
    disabled
    class="data-[highlighted]:font-bold"
    on:click={(e) => {
      e.preventDefault();
      isCustomRangeOpen = !isCustomRangeOpen;
    }}
  >
    Custom range
  </DropdownMenu.SubTrigger>

  <DropdownMenu.SubContent align="start" sideOffset={12}>
    <Calendar bind:start bind:end {zone} bind:selecting />
  </DropdownMenu.SubContent>
</DropdownMenu.Sub>

<form
  class="w-full overflow-hidden flex flex-col gap-y-6 p-2"
  id="custom-time-range-form"
  on:submit|preventDefault={handleSubmit}
>
  <div class="flex flex-col gap-2 size-full">
    <div class="date-wrapper">
      <label for="start-date" class="!font-medium">Start date (Inclusive)</label
      >
      <DateSelector
        selecting={isCustomRangeOpen && selecting === "start"}
        bind:value={start}
        maxYear={boundaryEnd.getFullYear()}
        minYear={boundaryStart.getFullYear()}
        {zone}
        label="start"
      />
    </div>

    <div class="date-wrapper">
      <label for="end-date" class="!font-medium">End date (Exclusive)</label>

      <DateSelector
        selecting={isCustomRangeOpen && selecting === "end"}
        bind:value={end}
        maxYear={boundaryEnd.getFullYear()}
        minYear={boundaryStart.getFullYear()}
        {zone}
        label="end"
      />
    </div>
  </div>
  {#if error}
    <div style:font-size="11px" class="text-red-600 mr-2">
      {error}
    </div>
  {/if}
  <div class="flex h-fit w-full justify-between">
    <div class="flex gap-x-1 items-center">
      <button
        aria-label="Nudge backward"
        on:click|preventDefault={() => {
          start = start.minus({ [nudgeGrain]: 1 });
          end = end.minus({ [nudgeGrain]: 1 });
        }}
        class="nudge rotate-180 pl-0.5"
      >
        <ChevronRight size="18px" />
      </button>
      <button
        aria-label="Nudge forward"
        class="nudge pl-0.5"
        on:click|preventDefault={() => {
          start = start.plus({ [nudgeGrain]: 1 });
          end = end.plus({ [nudgeGrain]: 1 });
        }}
      >
        <ChevronRight size="18px" />
      </button>
      <span>by</span>
      <div
        role="radiogroup"
        class="flex border rounded-full overflow-hidden"
        aria-label="Select time grain for date shift"
      >
        {#each ["day", "week", "month", "year"] as option}
          <input
            class="hidden"
            type="radio"
            id={option}
            bind:group={nudgeGrain}
            value={option}
          />
          <label
            for={option}
            class="cursor-pointer border-r !font-medium flex items-center px-1 last-of-type:border-r-0 last-of-type:pr-2 first-of-type:pl-2 hover:bg-primary-200"
            class:!bg-primary-200={nudgeGrain === option}
          >
            {option}
          </label>
        {/each}
      </div>
    </div>

    <Button {disabled} form="custom-time-range-form" submitForm type="primary">
      Apply
    </Button>
  </div>
</form>

<style lang="postcss">
  label {
    @apply font-semibold;
  }

  .date-wrapper {
    @apply flex flex-col w-full gap-1;
    @apply h-full;
  }

  .nudge {
    @apply flex items-center justify-center;
    @apply border border-gray-300 rounded-sm;
    @apply shadow-sm w-5 h-5;
  }

  .nudge:hover {
    @apply bg-gray-100;
  }

  .nudge:active {
    @apply bg-gray-200;
  }
</style>
