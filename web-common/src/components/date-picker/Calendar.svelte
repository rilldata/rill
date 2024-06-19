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

  function handlePan(direction: -1 | 1) {
    firstVisibleMonth = firstVisibleMonth.plus({ month: direction });
  }

  function handleSelectDay(date: DateTime<true>) {
    let newInterval: Interval;

    if (selectingStart) {
      newInterval = interval.set({ start: date });
    } else {
      newInterval = interval.set({ end: date.plus({ day: 1 }) });
    }

    if (newInterval.isValid) {
      interval = newInterval;
    } else {
      interval = Interval.fromDateTimes(
        date,
        date.plus({ day: 1 }),
      ) as Interval<true>;
    }
    selectingStart = !selectingStart;
  }
</script>

<div class="flex gap-x-3 p-2 w-full">
  {#each { length: visibleMonths } as month, i (month)}
    <Month
      {interval}
      startDay={firstVisibleMonth
        .plus({ month: i })
        .set({ day: 1 })
        .startOf("day")}
      {selectingStart}
      {visibleMonths}
      visibleIndex={i}
      bind:potentialStart
      bind:potentialEnd
      onSelectDay={handleSelectDay}
      onPan={handlePan}
    />
  {/each}
</div>
