<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { DateTime } from "luxon";
  import { ChevronLeft } from "lucide-svelte";
  import ChevronRight from "@rilldata/web-common/components/icons/ChevronRight.svelte";

  const dispatch = createEventDispatcher();

  export let month: DateTime;
  export let startOfWeek = 0;
  export let start: DateTime;
  export let end: DateTime;
  export let selecting: "start" | "end" = "start";
  export let zone: string;
  export let visibleMonths = 2;
  export let visibleIndex: number;
  export let potentialEnd: DateTime | undefined;
  export let potentialStart: DateTime | undefined;

  $: daysInMonth = month.daysInMonth ?? 0;
  $: firstDay = month.startOf("month").weekday;
  $: year = month.year;

  $: days = Array.from({ length: firstDay }, (_) => -1).concat(
    Array.from({ length: daysInMonth }, (_, i) => i + 1),
  );

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

<div class="flex flex-col gap-2">
  <div class="flex justify-between px-2">
    <button
      class:hide={visibleIndex !== 0}
      on:click={() => {
        dispatch("previous");
      }}
    >
      <ChevronLeft size="14px" />
    </button>
    <div class="w-full text-center p-2 text-sm flex gap-x-1 justify-center">
      <b>{month.monthLong}</b>
      <p>{month.year}</p>
    </div>
    <button
      class:hide={visibleIndex !== visibleMonths - 1}
      on:click={() => {
        dispatch("next");
      }}
    >
      <ChevronRight size="14px" />
    </button>
  </div>

  <div class="flex">
    {#each weekdays as day}
      <div class="weekday">{day}</div>
    {/each}
  </div>
  <div class="w-fit">
    <div class="grid grid-cols-7" class:selecting-start={selecting === "start"}>
      {#each days as day}
        {@const date = DateTime.fromObject({
          year,
          month: month.month,
          day,
        }).setZone(zone)}
        {@const inRange = date >= start && date <= end}
        {@const inPotentialRange =
          (potentialEnd && date > start && date < potentialEnd) ||
          (potentialStart && date > potentialStart && date < end)}

        {#if day > 0}
          <button
            class="flex items-center justify-center py-0.5"
            on:mouseleave={resetPotentialDates}
            on:click={() => {
              dispatch("select-day", date);
              resetPotentialDates();
            }}
            on:mouseenter={() => {
              if (selecting === "start") {
                potentialStart = date;
              } else {
                potentialEnd = date;
              }
            }}
          >
            <div
              class="day"
              class:in-range={inRange}
              class:in-potential-range={inPotentialRange}
              class:is-start={areSameDay(start, date)}
              class:is-end={areSameDay(end, date)}
            >
              {day}
            </div>
          </button>
        {:else}
          <div class="day pointer-events-none"></div>
        {/if}
      {/each}
    </div>
  </div>
</div>

<style lang="postcss">
  .day {
    @apply w-9 h-9;
    @apply p-0.5 bg-white;
    @apply flex items-center justify-center;
  }

  .in-range,
  .in-potential-range {
    @apply bg-gray-100;
  }

  .day:not(.is-start):not(.is-end):hover {
    @apply bg-gray-200 border-primary-500 border rounded-sm;
  }

  :not(.selecting-start) .day:not(.is-start):not(.is-end):hover {
    @apply rounded-sm border rounded-l-none border-l-0;
  }

  .selecting-start .day:not(.is-start):not(.is-end):hover {
    @apply rounded-sm border border-r-0 rounded-r-none;
  }

  .weekday {
    @apply text-center w-9 text-slate-500;
  }

  .in-potential-range {
    @apply border border-dashed border-primary-400 border-l-0 border-r-0;
  }

  .is-start,
  .is-end {
    @apply bg-primary-200;
  }

  .is-start {
    @apply rounded-l-sm;
  }

  .is-end {
    @apply rounded-r-sm;
  }

  .hide {
    @apply opacity-0 pointer-events-none;
  }
</style>
