<script lang="ts">
  import { DateTime, Interval } from "luxon";

  enum OverlapType {
    IN_RANGE = "in-range",
    START = "start",
    END = "end",
    FULL_INTERVAL = "full-interval",
  }

  export let date: DateTime<true>;
  export let interval: Interval<true>;
  export let outOfMonth: boolean;
  export let disabled: boolean;
  export let singleDaySelection: boolean = false;
  export let potentialInterval: Interval<true> | undefined = undefined;
  export let anchorDay: DateTime<true> | undefined;
  export let resetPotentialDates: () => void;
  export let onSelectDay: (date: DateTime<true>) => void;
  export let onHoverDay: (date: DateTime<true>) => void;

  $: overlapClass = getOverlapClass(date, potentialInterval, interval);

  function getOrderedInterval(
    date1: DateTime<true>,
    date2: DateTime<true>,
  ): Interval<true> | undefined {
    if (!date1 || !date2) return undefined;
    if (date1 < date2) {
      return Interval.fromDateTimes(
        date1.startOf("day"),
        date2.plus({ day: 1 }).startOf("day"),
      ) as Interval<true>;
    } else {
      return Interval.fromDateTimes(
        date2.startOf("day"),
        date1.plus({ day: 1 }).startOf("day"),
      ) as Interval<true>;
    }
  }

  function getOverlapClass(
    date: DateTime<true>,
    potentialInterval: Interval<true> | undefined,
    interval: Interval<true>,
  ): OverlapType | undefined {
    const evaluatedInterval = potentialInterval ?? interval;

    if (areSameDay(evaluatedInterval.start, date)) {
      if (areSameDay(evaluatedInterval.end.minus({ millisecond: 1 }), date))
        return OverlapType.FULL_INTERVAL;
      return OverlapType.START;
    }

    if (areSameDay(evaluatedInterval.end.minus({ millisecond: 1 }), date)) {
      return OverlapType.END;
    }

    if (evaluatedInterval.contains(date)) {
      return OverlapType.IN_RANGE;
    }

    return undefined;
  }

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
  class="py-0.5 wrapper size-full"
  {disabled}
  on:click={() => {
    onSelectDay(date);
    resetPotentialDates();
  }}
  on:mouseenter={() => {
    if (singleDaySelection) return;

    onHoverDay(date);

    if (!anchorDay) return;

    potentialInterval = getOrderedInterval(anchorDay, date);
  }}
>
  <div
    class="day {overlapClass}"
    class:text-gray-400={outOfMonth}
    class:potential={!!potentialInterval}
    class:anchor={areSameDay(anchorDay, date)}
  >
    {date.day}
  </div>
</button>

<style lang="postcss">
  .day {
    @apply w-full aspect-square;
    @apply bg-transparent;
    @apply flex flex-none items-center justify-center;
  }

  .day:disabled {
    @apply pointer-events-none text-gray-300;
  }

  button:hover .day:not(.potential) {
    @apply bg-primary-300 text-white;
  }

  button:hover .day:not(.in-range, .start, .end) {
    @apply rounded-full;
  }

  .in-range.potential,
  .start.potential,
  .end.potential,
  .full-interval.potential {
    @apply bg-gray-200 text-foreground;
  }

  .in-range,
  .start,
  .end,
  .full-interval,
  .anchor.potential {
    @apply bg-primary-400 text-white;
  }

  .end,
  .full-interval {
    @apply rounded-r-full;
  }

  .start,
  .full-interval {
    @apply rounded-l-full;
  }
</style>
