<script lang="ts">
  import { DateTime, Interval, type WeekdayNumbers } from "luxon";
  import { ChevronLeft } from "lucide-svelte";
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";
  import Day from "./Day.svelte";

  const weekCount = 6; // show 6 weeks to avoid layout shift

  export let interval: Interval<true>;
  export let startDay: DateTime<true>;
  export let firstDayOfWeek: WeekdayNumbers;
  export let visibleMonths = 2;
  export let visibleIndex: number;
  export let singleDaySelection = false;
  export let anchorDay: DateTime<true> | undefined;
  export let maxDate: DateTime<true> | DateTime<false> | undefined;
  export let minDate: DateTime<true> | DateTime<false> | undefined;
  export let onPan: (direction: 1 | -1) => void;
  export let onSelectDay: (date: DateTime<true>) => void;
  export let onHoverDay: (date: DateTime<true>) => void;

  let potentialInterval: Interval<true> | undefined = undefined;

  $: weekDayOfFirstDay = startDay.startOf("month").localWeekday;

  $: forwardPanEnabled = !maxDate || startDay.plus({ month: 1 }) < maxDate;
  $: backwardPanEnabled = !minDate || startDay.minus({ month: 1 }) >= minDate;

  $: days = Array.from({ length: weekCount * 7 }, (_, i) =>
    startDay.plus({ day: i + 1 - weekDayOfFirstDay }),
  );

  $: weekdays = Array.from({ length: 7 }, (_, i) => {
    return startDay.startOf("week").plus({ day: i + (firstDayOfWeek - 1) })
      .weekdayShort;
  });

  function resetPotentialDates() {
    potentialInterval = undefined;
  }
</script>

<div class="flex flex-col gap-2 w-full">
  <div class="flex justify-between items-center px-2">
    <button
      type="button"
      class="hover:bg-gray-200 rounded-full aspect-square size-5 flex items-center justify-center"
      class:opacity-50={!backwardPanEnabled}
      class:hide={visibleIndex !== 0}
      on:click={() => onPan(-1)}
    >
      <ChevronLeft size="14px" />
    </button>
    <div
      class="w-full text-center px-2 py-1 text-sm flex gap-x-1 justify-center select-none"
    >
      <b>{startDay.monthLong}</b>
      <p>{startDay.year}</p>
    </div>

    <button
      type="button"
      class="hover:bg-gray-200 rounded-full aspect-square size-5 flex items-center justify-center"
      class:opacity-50={!forwardPanEnabled}
      class:hide={visibleIndex !== visibleMonths - 1}
      on:click={() => onPan(1)}
    >
      <ChevronRight size="14px" />
    </button>
  </div>

  <div
    role="presentation"
    class="grid grid-cols-7 w-full"
    on:mouseleave={resetPotentialDates}
  >
    {#each weekdays as weekday (weekday)}
      <div class="weekday">{weekday}</div>
    {/each}
    {#each days as date (date.toISO())}
      <Day
        {date}
        {interval}
        bind:potentialInterval
        {singleDaySelection}
        {anchorDay}
        outOfMonth={date.month !== startDay.month}
        disabled={false}
        {onSelectDay}
        {resetPotentialDates}
        {onHoverDay}
      />
    {/each}
  </div>
</div>

<style lang="postcss">
  .weekday {
    @apply text-center w-full aspect-[2/1] text-slate-500 select-none;
  }

  .hide {
    @apply opacity-0 pointer-events-none;
  }
</style>
