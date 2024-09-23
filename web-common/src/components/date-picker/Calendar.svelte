<script lang="ts">
  import Month from "./Month.svelte";
  import { DateTime } from "luxon";
  import { Interval } from "luxon";

  export let dateSelection: DateTime<true> | undefined = undefined;
  export let rangeSelection: Interval | undefined = undefined;
  export let visibleMonths = 1;
  export let selectingStart = true;
  export let firstVisibleMonth: DateTime<true> =
    rangeSelection?.start ?? dateSelection ?? DateTime.now();
  export let singleDaySelection = dateSelection !== undefined;
  export let onSelectDay: (date: DateTime<true>) => void;

  let potentialEnd: DateTime | undefined;
  let potentialStart: DateTime | undefined;

  $: finalInterval = dateSelection
    ? (Interval.fromDateTimes(
        dateSelection,
        dateSelection.endOf("day"),
      ) as Interval<true>)
    : rangeSelection ?? Interval.invalid("No selection");

  function onPan(direction: -1 | 1) {
    firstVisibleMonth = firstVisibleMonth.plus({ month: direction });
  }
</script>

<div class="flex gap-x-3 p-2 w-full min-w-56">
  {#each { length: visibleMonths } as month, i (month)}
    <Month
      {singleDaySelection}
      interval={finalInterval}
      startDay={firstVisibleMonth
        .plus({ month: i })
        .set({ day: 1 })
        .startOf("day")}
      {selectingStart}
      {visibleMonths}
      visibleIndex={i}
      bind:potentialStart
      bind:potentialEnd
      {onSelectDay}
      {onPan}
    />
  {/each}
</div>
