<script lang="ts">
  import Month from "./Month.svelte";
  import type { DateTime } from "luxon";

  export let start: DateTime;
  export let end: DateTime;
  export let zone: string;
  export let visibleMonths = 2;
  export let selecting: "start" | "end" = "start";

  let potentialEnd: DateTime | undefined;
  let potentialStart: DateTime | undefined;

  let firstVisibleMonth = start;

  $: months = Array.from({ length: visibleMonths }, (_, i) =>
    firstVisibleMonth.plus({ month: i }),
  );

  function handleDecrementMonth() {
    firstVisibleMonth = firstVisibleMonth.plus({ month: -1 });
  }

  function handleIncrementMonth() {
    firstVisibleMonth = firstVisibleMonth.plus({ month: 1 });
  }

  function handleSelectDay(e: CustomEvent<DateTime>) {
    if (selecting === "start") {
      start = e.detail;
      selecting = "end";
    } else {
      end = e.detail;
      selecting = "start";
    }
  }
</script>

<div class="flex gap-x-3 p-1">
  {#each months as month, i}
    <Month
      on:select-day={handleSelectDay}
      on:previous={handleDecrementMonth}
      on:next={handleIncrementMonth}
      {month}
      {start}
      {selecting}
      {end}
      {zone}
      visibleIndex={i}
      bind:potentialStart
      bind:potentialEnd
    />
  {/each}
</div>
