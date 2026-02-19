<script lang="ts">
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { Interval } from "luxon";

  export let interval: Interval | undefined;
  export let timeGrain: V1TimeGrain = V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  export let abbreviation: string | undefined = undefined;

  $: formattedInterval = prettyFormatTimeRange(interval, timeGrain);
</script>

<div class="flex gap-x-1 whitespace-nowrap truncate">
  <span class="line-clamp-1 text-left">
    {#if interval?.isValid}
      {formattedInterval}
      {#if abbreviation}
        {abbreviation}
      {/if}
    {:else}
      Invalid Interval
    {/if}
  </span>
</div>
