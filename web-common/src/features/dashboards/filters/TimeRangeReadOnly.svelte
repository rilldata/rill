<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import { getRillTimeLabel } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
  import { getComparisonLabel } from "@rilldata/web-common/lib/time/comparisons";
  import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";

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
          {prettyFormatTimeRange(
            new Date(timeRange.start),
            new Date(timeRange.end),
            TimeRangePreset.CUSTOM,
            "UTC", // TODO
          )}
        </div>
      {:else if timeRange.expression}
        <div class:font-bold={hasBoldTimeRange}>
          {getRillTimeLabel(timeRange.expression)}
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
