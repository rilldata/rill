<script lang="ts">
  import type { DateTime } from "luxon";
  import { Interval } from "luxon";
  import Month from "./Month.svelte";

  export let interval: Interval<true>;
  export let visibleMonths = 1;
  export let selectingStart = true;
  export let firstVisibleMonth = interval.start;

  let potentialEnd: DateTime | undefined;
  let potentialStart: DateTime | undefined;

  $: months = Array.from({ length: visibleMonths }, (_, i) =>
    firstVisibleMonth.plus({ month: i }).set({ day: 1 }),
  );

  function handleDecrementMonth() {
    firstVisibleMonth = firstVisibleMonth.plus({ month: -1 });
  }

  function handleIncrementMonth() {
    firstVisibleMonth = firstVisibleMonth.plus({ month: 1 });
  }

  function handleSelectDay(e: CustomEvent<DateTime>) {
    const newInterval = interval.set({
      [selectingStart ? "start" : "end"]: e.detail,
    });

    if (newInterval.isValid) {
      interval = newInterval;
    } else {
      interval = Interval.fromDateTimes(
        e.detail,
        e.detail.plus({ day: 1 }),
      ) as Interval<true>;
    }
    selectingStart = !selectingStart;
  }
</script>

<div class="flex gap-x-3 p-2 w-full">
  {#each months as month, i (month)}
    <Month
      {interval}
      startDay={month}
      {selectingStart}
      {visibleMonths}
      visibleIndex={i}
      bind:potentialStart
      bind:potentialEnd
      on:select-day={handleSelectDay}
      on:previous={handleDecrementMonth}
      on:next={handleIncrementMonth}
    />
  {/each}
</div>
