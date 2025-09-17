<script lang="ts">
  import { DateTime } from "luxon";
  import SyntaxElement from "./SyntaxElement.svelte";
  import Timestamp from "./Timestamp.svelte";
  import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { V1TimeGrainToDateTimeUnit } from "@rilldata/web-common/lib/time/new-grains";

  export let min: DateTime | undefined;
  export let max: DateTime | undefined;
  export let timeZone: string;
  export let smallestTimeGrain: V1TimeGrain | undefined = undefined;

  const now = DateTime.now();

  $: unit = smallestTimeGrain
    ? V1TimeGrainToDateTimeUnit[smallestTimeGrain]
    : undefined;
</script>

<div
  class="bg-surface size-fit p-2 flex flex-col gap-y-1.5 border rounded-md shadow-md"
>
  {#if min}
    <div class="flex justify-between gap-x-3">
      <SyntaxElement range="earliest" />
      <Timestamp date={min} zone={timeZone} id="earliest" />
    </div>
  {/if}
  {#if max}
    <div class="flex justify-between gap-x-3">
      <SyntaxElement range="latest" />
      <Timestamp date={max} zone={timeZone} id="latest" />
    </div>
  {/if}
  <div class="flex justify-between gap-x-3">
    <SyntaxElement range="now" />
    <Timestamp date={now} zone={timeZone} id="now" />
  </div>

  {#if unit}
    <div class="text-xs text-gray-500 italic pt-1 text-center">
      Data granularity: {unit}
    </div>
  {/if}
</div>
