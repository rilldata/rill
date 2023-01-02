<script lang="ts">
  /** TimestampProfileSummary
   * ------------------------
   * This component provides summary information about the
   * timestamp profile at the top of the detail plot.
   */
  import Tooltip from "@rilldata/web-common/components/tooltip/Tooltip.svelte";
  import TooltipContent from "@rilldata/web-common/components/tooltip/TooltipContent.svelte";
  import { datesToFormattedTimeRange } from "@rilldata/web-common/lib/formatters";
  import GridCell from "@rilldata/web-local/lib/components/left-right-grid/GridCell.svelte";
  import LeftRightGrid from "@rilldata/web-local/lib/components/left-right-grid/LeftRightGrid.svelte";

  export let start: Date;
  export let end: Date;
  export let estimatedSmallestTimeGrain: string;
  export let rollupTimeGrain: string;

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

  $: formattedTimeRange = datesToFormattedTimeRange(start, end);

  $: displayRollupGrain = NicerTimeGrain[rollupTimeGrain];
</script>

<div class="ui-copy-muted" style:font-size="11px">
  <LeftRightGrid>
    <GridCell>
      <Tooltip distance={16} location="top">
        <div>
          {#if rollupTimeGrain}
            <span class="font-semibold">{formattedTimeRange}</span>
          {/if}
        </div>
        <TooltipContent slot="tooltip-content">
          <div style:max-width="315px">
            The range of this column is {formattedTimeRange}.
          </div>
        </TooltipContent>
      </Tooltip>
    </GridCell>
    <GridCell side="right">
      <Tooltip distance={16} location="top">
        <div>
          <span class="font-semibold">{displayRollupGrain}</span> row counts
        </div>

        <TooltipContent slot="tooltip-content">
          <div style:max-width="315px">
            This timestamp column is aggregated so each point on the time series
            represents a rollup count at the <b style:font-weight="600"
              >{displayRollupGrain} level</b
            >.
          </div>
        </TooltipContent>
      </Tooltip>
    </GridCell>
    <GridCell side="right">
      <Tooltip distance={16} location="top">
        <div class="text-right">
          {#if estimatedSmallestTimeGrain}
            min. interval at
            <span class="font-semibold"
              >{displayEstimatedSmallestTimegrain}</span
            >
            level
          {/if}
        </div>
        <TooltipContent slot="tooltip-content">
          <div style:max-width="315px">
            The smallest available time interval in this column appears to be at
            the <i>{displayEstimatedSmallestTimegrain}</i> level.
          </div>
        </TooltipContent>
      </Tooltip>
    </GridCell>
  </LeftRightGrid>
</div>
