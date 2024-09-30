<script context="module" lang="ts">
  import { Interval } from "luxon";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu/";
  import Calendar from "@rilldata/web-common/components/date-picker/Calendar.svelte";
  import { DateTime } from "luxon";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import DateInput from "@rilldata/web-common/components/date-picker/DateInput.svelte";
</script>

<script lang="ts">
  export let interval: Interval<true>;
  export let zone: string;
  export let firstVisibleMonth: DateTime<true>;
  export let applyRange: (range: Interval<true>) => void;
  export let closeMenu: () => void;

  let selectingStart = true;
  let displayError = false;

  $: calendarInterval = interval.set({
    start: interval.start.startOf("day"),
    end: interval.end.minus({ millisecond: 1 }).startOf("day"),
  });

  function onValidDateInput(date: DateTime) {
    let newInterval: Interval;

    const selectedEndDateBeforeStart =
      calendarInterval?.start && date < calendarInterval.start;

    if (selectingStart || selectedEndDateBeforeStart) {
      newInterval = calendarInterval.set({ start: date });
      selectingStart = false;
    } else {
      newInterval = calendarInterval.set({ end: date });
      selectingStart = true;
    }

    if (newInterval.isValid) {
      calendarInterval = newInterval;
    } else {
      const singleDay = Interval.fromDateTimes(date, date.endOf("day"));
      if (singleDay.isValid) {
        calendarInterval = singleDay;
      }
    }

    if (calendarInterval.isValid) {
      firstVisibleMonth = calendarInterval.start;
    }
  }
</script>

<svelte:window
  on:keydown|capture={(e) => {
    if (e.key === "Tab") {
      e.stopImmediatePropagation();
    }
  }}
/>

{#if calendarInterval.isValid}
  <Calendar
    selection={calendarInterval}
    {selectingStart}
    {firstVisibleMonth}
    onSelectDay={onValidDateInput}
  />
{/if}

<DropdownMenu.Separator />
<div class="flex flex-col gap-y-2 px-2 pt-1 pb-2">
  <label for="start-date" class:error={selectingStart && displayError}>
    Start Date
  </label>
  <div class="flex gap-x-1">
    <DateInput
      bind:selectingStart
      bind:displayError
      date={calendarInterval.start ?? DateTime.now()}
      {zone}
      boundary="start"
      currentYear={firstVisibleMonth.year}
      {onValidDateInput}
    />
  </div>
  <label for="start-date" class:error={!selectingStart && displayError}>
    End Date
  </label>

  <div class="flex gap-x-1 w-full">
    <DateInput
      bind:selectingStart
      bind:displayError
      date={calendarInterval.end ?? DateTime.now()}
      {zone}
      boundary="end"
      currentYear={firstVisibleMonth.year}
      {onValidDateInput}
    />
  </div>
</div>
<div class="flex justify-end w-full py-1 px-2">
  <Button
    fit
    compact
    type="primary"
    on:click={() => {
      const mapped = calendarInterval.set({
        end: calendarInterval.end?.plus({ day: 1 }).startOf("day"),
      });

      if (mapped.isValid) {
        applyRange(mapped);
      }

      closeMenu();
    }}
  >
    <span class="px-2 w-fit">Apply</span>
  </Button>
</div>

<style lang="postcss">
  label {
    @apply font-semibold flex gap-x-1;
  }
</style>
