<script lang="ts">
  import { DateTime } from "luxon";
  import SyntaxElement from "./SyntaxElement.svelte";
  import Timestamp from "./Timestamp.svelte";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";

  export let timeStart: Date | undefined;
  export let timeEnd: Date | undefined;
  export let timeZone: string;
  export let smallestTimeGrain: V1TimeGrain | undefined;

  const now = DateTime.now();

  // $: earliest =

  // $: adjustedEnd
</script>

<div
  class="bg-surface size-fit p-2 flex z-100 flex-col gap-y-1.5 border rounded-[2px] shadow-md"
>
  {#if timeStart}
    <div class="flex justify-between gap-x-3">
      <SyntaxElement range="earliest" />
      <Timestamp date={DateTime.fromJSDate(timeStart)} zone={timeZone} />
    </div>
  {/if}
  {#if timeEnd}
    <div class="flex justify-between gap-x-3">
      <SyntaxElement range="latest" />
      <Timestamp date={DateTime.fromJSDate(timeEnd)} zone={timeZone} />
    </div>
  {/if}
  {#if smallestTimeGrain && timeEnd}
    <div class="flex justify-between gap-x-3">
      <SyntaxElement range="now" />
      <Timestamp
        date={DateTime.fromJSDate(timeEnd, { zone: timeZone })
          .plus({ [V1TimeGrainToDateTimeUnit[smallestTimeGrain]]: 1 })
          .startOf(V1TimeGrainToDateTimeUnit[smallestTimeGrain])}
        zone={timeZone}
      />
    </div>
  {/if}
  <!-- <div class="flex justify-between gap-x-3">
    <SyntaxElement range="wallclock" />
    <Timestamp date={now} zone={timeZone} />
  </div> -->
</div>
