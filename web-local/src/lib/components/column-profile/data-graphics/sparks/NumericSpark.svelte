<script lang="ts">
  import { COLUMN_PROFILE_CONFIG } from "@rilldata/web-local/lib/application-config";
  import { DATA_TYPE_COLORS } from "@rilldata/web-local/lib/duckdb-data-types";
  import Tooltip from "../../../tooltip/Tooltip.svelte";
  import TooltipContent from "../../../tooltip/TooltipContent.svelte";
  import Histogram from "../../../viz/histogram/SmallHistogram.svelte";

  export let compact = false;
  export let data;

  $: summaryWidthSize =
    COLUMN_PROFILE_CONFIG.summaryVizWidth[compact ? "small" : "medium"];
</script>

{#if data}
  <Tooltip location="right" alignment="center" distance={8}>
    <Histogram
      {data}
      width={summaryWidthSize}
      height={18}
      fillColor={DATA_TYPE_COLORS["DOUBLE"].vizFillClass}
      baselineStrokeColor={DATA_TYPE_COLORS["DOUBLE"].vizStrokeClass}
    />
    <TooltipContent slot="tooltip-content">
      the distribution of the values of this column
    </TooltipContent>
  </Tooltip>
{/if}
