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
  export let applyRange: (range: Interval<true>) => void;
  export let firstVisibleMonth: DateTime<true>;
  export let closeMenu: () => void;

  let selectingStart = true;
  let displayError = false;

  function onValidDateInput(date: DateTime) {
    const newInterval = interval.set({
      [selectingStart ? "start" : "end"]: date,
    });

    if (newInterval.isValid) {
      interval = newInterval;
    } else {
      const singleDay = Interval.fromDateTimes(date, date.plus({ day: 1 }));
      if (singleDay.isValid) {
        interval = singleDay;
      }
    }

    firstVisibleMonth = interval.start;
  }
</script>

<Calendar bind:interval bind:selectingStart bind:firstVisibleMonth />
<DropdownMenu.Separator />
<div class="flex flex-col gap-y-2 px-2 pt-1 pb-2">
  <label for="start-date" class:error={selectingStart && displayError}>
    Start Date <span class="secondary">(Inclusive)</span>
  </label>
  <div class="flex gap-x-1">
    <DateInput
      bind:selectingStart
      bind:displayError
      date={interval.start}
      {zone}
      boundary="start"
      currentYear={firstVisibleMonth.year}
      {onValidDateInput}
    />
  </div>
  <label for="start-date" class:error={!selectingStart && displayError}>
    End Date <span class="secondary">(Exclusive)</span>
  </label>
  <div class="flex gap-x-1 w-full">
    <DateInput
      bind:selectingStart
      bind:displayError
      date={interval.end}
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
    on:click={() => {
      const mapped = interval.mapEndpoints((dt) => dt.startOf("day"));
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

  /* label.error {
      @apply text-destructive;
    } */
</style>
