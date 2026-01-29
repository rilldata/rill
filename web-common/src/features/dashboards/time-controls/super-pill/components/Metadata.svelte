<script lang="ts">
  import { DateTime } from "luxon";
  import SyntaxElement from "./SyntaxElement.svelte";
  import Timestamp from "./Timestamp.svelte";

  export let timeStart: Date | undefined;
  export let timeEnd: Date | undefined;
  export let timeZone: string;

  const now = DateTime.now();
</script>

<div
  class="bg-tooltip text-fg-inverse size-fit p-2 flex flex-col gap-y-1.5 rounded-md shadow-md"
>
  {#if timeStart}
    <div class="flex justify-between gap-x-3">
      <SyntaxElement dark range="earliest" />
      <Timestamp
        date={DateTime.fromJSDate(timeStart)}
        zone={timeZone}
        id="earliest"
      />
    </div>
  {/if}
  {#if timeEnd}
    <div class="flex justify-between gap-x-3">
      <SyntaxElement dark range="latest" />
      <Timestamp
        date={DateTime.fromJSDate(timeEnd)}
        zone={timeZone}
        id="latest"
      />
    </div>
  {/if}
  <div class="flex justify-between gap-x-3">
    <SyntaxElement dark range="now" />
    <Timestamp date={now} zone={timeZone} id="now" />
  </div>
</div>
