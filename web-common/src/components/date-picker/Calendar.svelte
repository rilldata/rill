<script lang="ts">
  import Month from "./Month.svelte";
  import type { DateTime } from "luxon";

  export let start: DateTime;
  export let end: DateTime;
  export let zone: string;
  export let visibleMonths = 2;
  export let selectingStart: boolean;

  let potentialEnd: DateTime | undefined;
  let potentialStart: DateTime | undefined;

  let firstVisibleMonth = start;

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
    if (selectingStart) {
      start = e.detail;
      selectingStart = false;
    } else {
      end = e.detail;
      selectingStart = true;
    }
  }
</script>

<div class="flex gap-x-3 p-1">
  {#each months as month, i}
    <Month
      {start}
      {end}
      {zone}
      startDay={month}
      {selectingStart}
      visibleIndex={i}
      bind:potentialStart
      bind:potentialEnd
      on:select-day={handleSelectDay}
      on:previous={handleDecrementMonth}
      on:next={handleIncrementMonth}
    />
  {/each}
</div>
