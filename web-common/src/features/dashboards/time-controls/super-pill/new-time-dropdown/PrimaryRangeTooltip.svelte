<script lang="ts">
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { Interval, DateTime } from "luxon";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let timeString: string | undefined;
  export let interval: Interval<true>;
  // Set only when the active time axis is a time dimension (not a raw column).
  export let timeDimensionLabel: string | undefined = undefined;
  export let timeDimensionDescription: string | undefined = undefined;
</script>

<TooltipContent class="flex-col flex items-center gap-y-0 p-3">
  {#if timeDimensionLabel}
    <div class="flex flex-col items-center mb-1">
      <span class="font-bold text-fg-inverse">{timeDimensionLabel}</span>
      {#if timeDimensionDescription}
        <span class="text-fg-inverse/70 text-center max-w-64">
          {timeDimensionDescription}
        </span>
      {/if}
    </div>
  {/if}
  <span class="font-semibold italic mb-1">{timeString}</span>
  {#if interval.isValid}
    <span
      >{interval.start.toLocaleString({
        ...DateTime.DATETIME_HUGE_WITH_SECONDS,
        fractionalSecondDigits: interval.start.millisecond > 0 ? 3 : undefined,
        second: interval.start.second > 0 ? "numeric" : undefined,
      })}
    </span>
    <span>{m.dashboard_to()}</span>
    <span
      >{interval.end.toLocaleString({
        ...DateTime.DATETIME_HUGE_WITH_SECONDS,
        fractionalSecondDigits: interval.end.millisecond > 0 ? 3 : undefined,
        second: interval.end.second > 0 ? "numeric" : undefined,
      })}
    </span>
  {:else}
    <span class="text-fg-secondary">{m.dashboard_invalid_time_range()}</span>
  {/if}
</TooltipContent>
