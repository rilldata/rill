<script lang="ts">
  /** TimestampProfileSummary
   * ------------------------
   * This component provides summary information about the
   * timestamp profile at the top of the detail plot.
   */
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import type { Interval } from "@rilldata/web-common/lib/duckdb-data-types";
  import {
    intervalToTimestring,
    PreviewRollupIntervalFormatter,
  } from "@rilldata/web-common/lib/formatters";

  export let type: string;
  export let estimatedSmallestTimeGrain: string;
  export let interval: Interval;
  export let rollupGrain: string;

  enum NicerTimeGrain {
    TIME_GRAIN_MILLISECOND = "milliseconds",
    TIME_GRAIN_SECOND = "seconds",
    TIME_GRAIN_MINUTE = "minutes",
    TIME_GRAIN_HOUR = "hourly",
    TIME_GRAIN_DAY = "daily",
    TIME_GRAIN_WEEK = "weekly",
    TIME_GRAIN_MONTH = "monthly",
    TIME_GRAIN_YEAR = "yearly",
  }

  $: displayEstimatedSmallestTimegrain =
    NicerTimeGrain?.[estimatedSmallestTimeGrain] || estimatedSmallestTimeGrain;
</script>

<div
  class="text-gray-500 pb-3"
  style="
        display: grid;
        grid-template-columns: auto auto;
    "
>
  <Tooltip distance={16} location="top">
    <div style:font-weight="600">
      {type}
    </div>

    <TooltipContent slot="tooltip-content">
      <div style:max-width="315px">
        This column has the {type} type.
      </div>
    </TooltipContent>
  </Tooltip>

  <Tooltip distance={16} location="top">
    <div class="text-right">
      {#if estimatedSmallestTimeGrain}
        min. interval at
        {displayEstimatedSmallestTimegrain}
        level
      {/if}
    </div>
    <TooltipContent slot="tooltip-content">
      <div style:max-width="315px">
        The smallest available time interval in this column appears to be at the <i
          >{displayEstimatedSmallestTimegrain}</i
        > level.
      </div>
    </TooltipContent>
  </Tooltip>

  <Tooltip distance={16} location="top">
    <div>
      {#if interval}
        {intervalToTimestring(interval)}
      {/if}
    </div>
    <TooltipContent slot="tooltip-content">
      <div style:max-width="315px">
        The range of this timestamp is {intervalToTimestring(interval)}.
      </div>
    </TooltipContent>
  </Tooltip>

  <Tooltip distance={16} location="top">
    <div class="text-right">
      {#if rollupGrain}
        showing {PreviewRollupIntervalFormatter[rollupGrain]} row counts
      {/if}
    </div>
    <TooltipContent slot="tooltip-content">
      <div style:max-width="315px">
        This timestamp column is aggregated so each point on the time series
        represents a rollup count at the <b style:font-weight="600"
          >{NicerTimeGrain?.[rollupGrain] || "unknown"} level</b
        >.
      </div>
    </TooltipContent>
  </Tooltip>
</div>
