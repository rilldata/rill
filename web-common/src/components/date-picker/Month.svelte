<script lang="ts">
  import type { DateTime, Interval } from "luxon";
  import { ChevronLeft } from "lucide-svelte";
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";

  export let startDay: DateTime<true>;
  export let startOfWeek = 0;
  export let interval: Interval<true>;
  export let selectingStart: boolean;
  export let visibleMonths = 2;
  export let visibleIndex: number;
  export let potentialEnd: DateTime | undefined;
  export let potentialStart: DateTime | undefined;
  export let onPan: (direction: 1 | -1) => void;
  export let onSelectDay: (date: DateTime<true>) => void;

  $: firstDay = startDay.startOf("month").weekday % 7;

  $: weekCount = Math.ceil((firstDay + startDay.daysInMonth) / 7);

  $: inclusiveEnd = interval.end.minus({ day: 1 });

  $: days = Array.from({ length: weekCount * 7 }, (_, i) => {
    if (i < firstDay) {
      return startDay.minus({ day: firstDay - i });
    } else {
      return startDay.plus({ day: i - firstDay });
    }
  });

  $: weekdays = Array.from({ length: 7 }, (_, i) =>
    new Date(0, 0, i + startOfWeek).toLocaleString("default", {
      weekday: "short",
    }),
  );

  function resetPotentialDates() {
    potentialEnd = undefined;
    potentialStart = undefined;
  }

  function areSameDay(a: DateTime, b: DateTime) {
    return a.hasSame(b, "day") && a.hasSame(b, "month") && a.hasSame(b, "year");
  }
</script>

<div class="flex flex-col gap-2 w-full">
  <div class="flex justify-between px-2">
    <button
      class:hide={visibleIndex !== 0}
      class="hover:opacity-50"
      on:click={() => {
        onPan(-1);
      }}
    >
      <ChevronLeft size="14px" />
    </button>
    <div
      class="w-full text-center px-2 py-1 text-sm flex gap-x-1 justify-center"
    >
      <b>{startDay.monthLong}</b>
      <p>{startDay.year}</p>
    </div>
    <button
      class="hover:opacity-50"
      class:hide={visibleIndex !== visibleMonths - 1}
      on:click={() => {
        onPan(1);
      }}
    >
      <ChevronRight size="14px" />
    </button>
  </div>

  <div
    role="presentation"
    class="grid grid-cols-7 gap-y-1 w-full"
    class:selecting-start={selectingStart}
    on:mouseleave={resetPotentialDates}
  >
    {#each weekdays as weekday (weekday)}
      <div class="weekday">{weekday}</div>
    {/each}
    {#each days as date (date.toISO())}
      {@const isEnd = areSameDay(inclusiveEnd, date)}
      {@const inRange =
        !isEnd && date >= interval.start && date <= inclusiveEnd}
      {@const inPotentialRange =
        (potentialEnd && date > interval.start && date < potentialEnd) ||
        (potentialStart && date > potentialStart && date < inclusiveEnd)}
      {@const outOfMonth = date.month !== startDay.month}
      {@const weekend = date.weekday === 6 || date.weekday === 7}
      <button
        class="day font-medium"
        on:click={() => {
          onSelectDay(date);
          resetPotentialDates();
        }}
        on:mouseenter={() => {
          if (selectingStart) {
            potentialStart = date;
          } else {
            potentialEnd = date;
          }
        }}
        class:!font-normal={outOfMonth}
        class:text-gray-500={weekend}
        class:!text-gray-300={outOfMonth}
        class:in-range={inRange}
        class:in-potential-range={inPotentialRange}
        class:is-start={areSameDay(interval.start, date)}
        class:is-end={isEnd}
      >
        {date.day}
      </button>
    {/each}
  </div>
</div>

<style lang="postcss">
  .day {
    @apply w-full aspect-square;
    @apply p-0.5 bg-transparent;
    @apply flex items-center justify-center border border-transparent border-l-0 border-r-0;
  }

  :not(.selecting-start) .day:not(.is-start):not(.is-end):hover {
    @apply rounded-r-sm  border border-primary-200 bg-primary-200 border-dashed;
  }

  .selecting-start .day:not(.is-start):not(.is-end):hover {
    @apply rounded-l-md;
    @apply bg-primary-600 border-primary-600 border-l border-r-0 text-white;
  }

  .weekday {
    @apply text-center w-full aspect-[2/1] text-slate-500;
  }

  .hide {
    @apply opacity-0 pointer-events-none;
  }

  .is-end {
    @apply rounded-r-md border border-l-0;
  }

  .in-range {
    @apply bg-primary-100 border-primary-200;
  }

  :not(.in-range).in-potential-range {
    @apply bg-primary-50 border-dashed border-primary-200;
  }

  .is-start {
    @apply rounded-l-md;
    @apply border-l border-r-0;
  }

  .is-end {
    @apply bg-primary-700 border-primary-700 text-white;
  }

  .is-start {
    @apply bg-primary-600 border-primary-600 text-white;
  }
</style>
