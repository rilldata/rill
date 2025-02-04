<script lang="ts">
  import { DateTime } from "luxon";

  export let date: DateTime<true>;
  export let inclusiveEnd: DateTime<true> | undefined;
  export let start: DateTime<true> | undefined;
  export let outOfMonth: boolean;
  export let disabled: boolean;
  export let selectingStart: boolean;
  export let singleDaySelection: boolean;
  export let potentialStart: DateTime | undefined;
  export let potentialEnd: DateTime | undefined;
  export let resetPotentialDates: () => void;
  export let onSelectDay: (date: DateTime<true>) => void;

  $: isEnd = areSameDay(inclusiveEnd, date);
  $: isStart = areSameDay(start, date);
  $: afterEnd = inclusiveEnd && date > inclusiveEnd;
  $: beforeStart = start && date < start;
  $: inPotentialRange = Boolean(
    (start && potentialEnd && date > start && date < potentialEnd) ||
      (inclusiveEnd &&
        potentialStart &&
        date > potentialStart &&
        date < inclusiveEnd),
  );
  $: isNextDay = areSameDay(potentialStart?.plus({ day: 1 }), date);
  $: inRange = start && inclusiveEnd && !afterEnd && !beforeStart;

  function areSameDay(
    a: DateTime | undefined | null,
    b: DateTime | undefined | null,
  ) {
    if (!a || !b) return false;
    return a.hasSame(b, "day") && a.hasSame(b, "month") && a.hasSame(b, "year");
  }
</script>

<button
  type="button"
  class="day"
  {disabled}
  class:text-gray-400={outOfMonth ||
    (!singleDaySelection && beforeStart && !selectingStart)}
  class:in-range={inRange}
  class:in-potential-range={inPotentialRange}
  class:is-start={isStart}
  class:is-end={isEnd}
  class:before-start={beforeStart}
  class:after-end={afterEnd}
  class:end-cap={!selectingStart && !beforeStart}
  class:start-cap={selectingStart || (!selectingStart && beforeStart)}
  class:next-day={isNextDay}
  class:single-day-selection={singleDaySelection}
  on:click={() => {
    onSelectDay(date);
    resetPotentialDates();
  }}
  on:mouseleave={resetPotentialDates}
  on:mouseenter={() => {
    if (singleDaySelection) return;

    if (selectingStart || (start && date < start)) {
      potentialStart = date;
    } else {
      potentialEnd = date;
    }
  }}
>
  {date.day}
</button>

<style lang="postcss">
  .day {
    @apply font-medium;
    @apply w-full aspect-[5/4];
    @apply p-0.5 bg-transparent;
    @apply flex items-center justify-center;
    @apply border border-transparent border-l-0 border-r-0;
  }

  .day:disabled {
    @apply pointer-events-none text-gray-300;
  }

  .day:hover {
    @apply bg-primary-600;
    @apply border-primary-600;
    color: white !important;
  }

  .end-cap:hover {
    @apply rounded-r-md;
  }

  .start-cap:hover {
    @apply rounded-l-md;
  }

  .is-end {
    @apply rounded-r-full border border-l-0;
  }

  .in-range {
    @apply bg-primary-50 border-primary-200;
  }

  :not(.in-range).in-potential-range {
    @apply bg-primary-50 border-dashed border-primary-200;
  }

  .is-end {
    @apply bg-primary-100 border-primary-200;
  }

  .is-start {
    @apply bg-primary-100 border-primary-200;
    @apply rounded-l-full;
    @apply border-l border-r-0;
  }

  .single-day-selection {
    @apply rounded-md;
  }

  :not(.single-day-selection).next-day:not(.in-potential-range) {
    @apply border-dashed border-t border-b border-primary-200;
    @apply bg-gradient-to-r from-primary-100 to-transparent;
  }
</style>
