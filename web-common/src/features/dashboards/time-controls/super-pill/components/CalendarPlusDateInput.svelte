<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Calendar from "@rilldata/web-common/components/date-picker/Calendar.svelte";
  import DateInput from "@rilldata/web-common/components/date-picker/DateInput.svelte";
  import { DateTime, Interval, type DateTimeUnit } from "luxon";

  export let interval: Interval<true> | undefined;
  export let minDate: DateTime | undefined = undefined;
  export let maxDate: DateTime | undefined = undefined;
  export let minTimeGrain: DateTimeUnit;
  export let zone: string;
  export let updateRange: (range: string) => void = () => {};
  export let onApply: (interval: Interval<true>) => void;
  export let closeMenu: () => void;

  const now = DateTime.now().setZone(zone);
  const today: Interval<true> = Interval.fromDateTimes(
    now.startOf("day"),
    now.plus({ day: 1 }).startOf("day"),
  ) as Interval<true>;

  let inputInterval = interval || today;
  let firstVisibleMonth = inputInterval.start;
  let anchorDay: DateTime<true> | undefined = undefined;

  $: startDate = inputInterval?.start;
  $: endDate = inputInterval.end.minus({ millisecond: 1 });

  $: adjustedMinDate = minDate?.startOf("day");
  $: adjustedMaxDate = maxDate
    ?.plus({ [minTimeGrain]: 1 })
    .startOf(minTimeGrain);

  function onValidDateInput(date: DateTime<true>, boundary?: "start" | "end") {
    let newInterval: Interval;

    if (boundary) {
      if (boundary === "start") {
        newInterval = Interval.fromDateTimes(date, inputInterval.end);
      } else {
        newInterval = Interval.fromDateTimes(inputInterval.start, date);
      }
    } else if (!anchorDay) {
      anchorDay = date;
      newInterval = Interval.fromDateTimes(
        anchorDay,
        anchorDay.plus({ day: 1 }).startOf("day"),
      ) as Interval<true>;
    } else if (date > anchorDay) {
      newInterval = Interval.fromDateTimes(
        anchorDay,
        date.plus({ day: 1 }).startOf("day"),
      ) as Interval<true>;
      anchorDay = undefined;
    } else {
      newInterval = Interval.fromDateTimes(
        date.startOf("day"),
        anchorDay.plus({ day: 1 }).startOf("day"),
      ) as Interval<true>;
      anchorDay = undefined;
    }

    if (newInterval.isValid) {
      inputInterval = newInterval;
    } else {
      const singleDay = Interval.fromDateTimes(date, date.endOf("day"));
      if (singleDay.isValid) {
        inputInterval = singleDay;
      }
    }

    updateRange(
      `${inputInterval.start.toFormat("yyyy-MM-dd")} to ${inputInterval.end.toFormat("yyyy-MM-dd")}`,
    );
  }
</script>

<svelte:window
  on:keydown|capture={(e) => {
    if (e.key === "Tab") {
      e.stopImmediatePropagation();
    }
  }}
/>

<div class="flex flex-col w-full gap-y-3">
  <Calendar
    minDate={adjustedMinDate}
    maxDate={adjustedMaxDate}
    selection={inputInterval}
    {anchorDay}
    {firstVisibleMonth}
    onSelectDay={onValidDateInput}
  />

  <!-- {#if usingRillTime} -->
  <div class="w-full h-px bg-gray-200"></div>

  <div class="flex flex-col gap-y-2">
    <DateInput
      boundary="start"
      {zone}
      minDate={adjustedMinDate}
      maxDate={adjustedMaxDate}
      currentYear={firstVisibleMonth.year}
      date={startDate}
      {onValidDateInput}
      onFocus={() => {
        firstVisibleMonth = inputInterval.start;
      }}
    />

    <DateInput
      boundary="end"
      {zone}
      date={endDate}
      minDate={adjustedMinDate}
      maxDate={adjustedMaxDate}
      currentYear={firstVisibleMonth.year}
      {onValidDateInput}
      onFocus={() => {
        firstVisibleMonth = inputInterval.end;
      }}
    />
  </div>
  <!-- {/if} -->
  <div class="flex justify-end w-full">
    <Button
      fit
      compact
      type="subtle"
      disabled={!inputInterval?.isValid}
      onClick={() => {
        onApply(inputInterval);

        closeMenu();
      }}
    >
      <span class="px-2 w-fit">Apply</span>
    </Button>
  </div>
</div>
