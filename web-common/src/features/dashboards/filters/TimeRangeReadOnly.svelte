<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import { getComparisonLabel } from "@rilldata/web-common/lib/time/comparisons";
  import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";
  import { isValidDateTime } from "@rilldata/web-common/lib/time/is-valid-datetime";

  export let timeRange: V1TimeRange;
  export let comparisonTimeRange: V1TimeRange | undefined;
  export let hasBoldTimeRange: boolean = true;
</script>

<Chip type="time" readOnly>
  <svelte:fragment slot="body">
    <div class="text-xs text-slate-800 px-2">
      {#if timeRange.isoDuration && timeRange.isoDuration !== TimeRangePreset.CUSTOM}
        <div class="font-bold">
          {#if timeRange.isoDuration === TimeRangePreset.ALL_TIME}
            All Time
          {:else if timeRange.isoDuration in DEFAULT_TIME_RANGES}
            {DEFAULT_TIME_RANGES[timeRange.isoDuration].label}
          {:else}
            Last {humaniseISODuration(timeRange.isoDuration)}
          {/if}
        </div>
      {:else if timeRange.start && timeRange.end}
        <div class:font-bold={hasBoldTimeRange}>
          {#if isValidDateTime(timeRange.start) && isValidDateTime(timeRange.end)}
            {prettyFormatTimeRange(
              new Date(timeRange.start),
              new Date(timeRange.end),
              TimeRangePreset.CUSTOM,
              "UTC", // TODO
            )}
          {:else}
            -
          {/if}
        </div>
      {/if}
    </div>
  </svelte:fragment>
</Chip>

{#if comparisonTimeRange}
  <Chip type="time" readOnly>
    <svelte:fragment slot="body">
      <div class="text-xs text-slate-800 px-2">
        vs
        <span class:font-bold={hasBoldTimeRange}>
          {getComparisonLabel(comparisonTimeRange)}
        </span>
      </div>
    </svelte:fragment>
  </Chip>
{/if}
