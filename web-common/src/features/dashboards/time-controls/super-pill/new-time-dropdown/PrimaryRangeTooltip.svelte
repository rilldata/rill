<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { Interval, DateTime } from "luxon";

  export let timeString: string | undefined;
  export let interval: Interval<true>;
</script>

<TooltipContent class="flex-col flex items-center gap-y-0 p-3">
  <span class="font-semibold italic mb-1">{timeString}</span>
  {#if interval.isValid}
    <span
      >{interval.start.toLocaleString({
        ...DateTime.DATETIME_HUGE_WITH_SECONDS,
        fractionalSecondDigits: interval.start.millisecond > 0 ? 3 : undefined,
        second: interval.start.second > 0 ? "numeric" : undefined,
      })}
    </span>
    <span>to</span>
    <span
      >{interval.end.toLocaleString({
        ...DateTime.DATETIME_HUGE_WITH_SECONDS,
        fractionalSecondDigits: interval.end.millisecond > 0 ? 3 : undefined,
        second: interval.end.second > 0 ? "numeric" : undefined,
      })}
    </span>
  {:else}
    <span class="text-gray-500">Invalid time range</span>
  {/if}
</TooltipContent>
