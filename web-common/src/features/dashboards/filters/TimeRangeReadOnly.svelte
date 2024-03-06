<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import { timeChipColors } from "@rilldata/web-common/components/chip/chip-types";
  import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
  import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges";
  import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
  import { TimeRangePreset } from "@rilldata/web-common/lib/time/types";
  import type { V1TimeRange } from "@rilldata/web-common/runtime-client";

  export let timeRange: V1TimeRange;
</script>

<Chip {...timeChipColors} outline readOnly>
  <svelte:fragment slot="body">
    <div class="font-bold text-xs text-slate-800 px-2">
      {#if timeRange.isoDuration && timeRange.isoDuration !== TimeRangePreset.CUSTOM}
        {#if timeRange.isoDuration === TimeRangePreset.ALL_TIME}
          All Time
        {:else if timeRange.isoDuration in DEFAULT_TIME_RANGES}
          {DEFAULT_TIME_RANGES[timeRange.isoDuration].label}
        {:else}
          Last {humaniseISODuration(timeRange.isoDuration)}
        {/if}
      {:else if timeRange.start && timeRange.end}
        {prettyFormatTimeRange(
          new Date(timeRange.start),
          new Date(timeRange.end),
          TimeRangePreset.CUSTOM,
          "UTC", // TODO
        )}
      {/if}
    </div>
  </svelte:fragment>
</Chip>
