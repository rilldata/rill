<script lang="ts">
  import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
  import { Interval, DateTime } from "luxon";

  const showHour = [
    "TIME_GRAIN_HOUR",
    "TIME_GRAIN_MINUTE",
    "TIME_GRAIN_SECOND",
  ];

  export let interval: Interval<true>;
  export let grain: V1TimeGrain | undefined;

  $: date = interval.toLocaleString(DateTime.DATE_MED);

  $: time = interval.toFormat("h a", { separator: "-" });
</script>

<div class="flex gap-x-1">
  <span class="line-clamp-1">{date}</span>

  {#if grain && showHour.includes(grain)}
    <span class="line-clamp-1">({time})</span>
  {/if}
</div>
