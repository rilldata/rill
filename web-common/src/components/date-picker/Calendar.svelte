<script lang="ts">
  import { DateTime } from "luxon";
  import { Interval } from "luxon";
  import Month from "./Month.svelte";

  export let interval: Interval;
  export let visibleMonths = 1;
  export let selectingStart = true;
  export let firstVisibleMonth = interval.start ?? DateTime.now();
  export let onSelectDay: (date: DateTime<true>) => void;

  let potentialEnd: DateTime | undefined;
  let potentialStart: DateTime | undefined;

  function onPan(direction: -1 | 1) {
    firstVisibleMonth = firstVisibleMonth.plus({ month: direction });
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
      {onSelectDay}
      {onPan}
    />
  {/each}
</div>
